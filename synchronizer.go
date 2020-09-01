package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
)

// Synchronizer matém as configuraçõse do programa
type Synchronizer struct {
	autoFind        bool            // Marca se o modo utilizado é o de localizar automaticamente as pastas do caso
	sources         []string        // pastas fontes
	source          string          // Pasta fonte corrente
	caseName        string          // Nome do caso
	dest            string          // pasta destino
	verbose         bool            // Imprimir mensagens extras
	NWorkers        int             // Número de workers
	maxDepth        int             // Profundida máxia para busca de pastas de caso
	threshold       int64           // Tamanho em megabytes a partir do qual será copiado sem concorrência
	thresholdChunk  int64           // Tamanho acima do qual o arquivo será copiado em partes
	bufferSize      int64           // Tamanho do buffer utilizado para copiar arquivos grandes
	purge           bool            // Deletar o que existir no destino e não na fonte
	retries         int             // Número de tentativas de copiar arquivo caso ocorra erros
	wait            time.Duration   // Tempo para aguardar antes de tentar de novo em segundos
	results         chan ResultData // Canal cuja função é ser utilizado pelos workers para passar os resultados das tarefas ao gerenciador de progresso
	jobs            chan JobConfig  // Canal cuja função é ser utilizado pelo sincronizador para atribuir tarefas aos workers
	acknowledgeDone chan bool       // Canal cuja função é ser utilizado pelos workers para informar o sincronizador que uma cópia de arquivo grande foi finalizada
	finished        chan bool       // Canal cuja função é ser utilizado pelo gerenciador de progresso para informar o sincronizador que o trabalho de sincronização terminou
	wg              sync.WaitGroup  // Este objeto é utlizado pelo sincronizador para garantir que não tem nem um workers trabalhando ainda
	logger          Logger          //utilizado para usar o canal de comunicação do logger
	logging         bool
}

func (synchronizer *Synchronizer) init() {
	synchronizer.jobs = make(chan JobConfig)
	synchronizer.results = make(chan ResultData)
	synchronizer.finished = make(chan bool)
}

// Função de execução do sincronizador. Aqui as tarefas vão sendo colocadas no canal jobs uma a uma.
// A medida que cada woker vai terminando sua tarefa ele vai pegando novas tarefas no canal jobs.
// O sincronzador coloca uma tarefa e espera até que algum worker a pegue pra fazer para só então colocar a próxima no canal
func (synchronizer *Synchronizer) run() {

	synchronizer.acknowledgeDone = make(chan bool)
	//não devia estar no init?

	//Cria a pasta de destino caso ela não exista
	_, err := os.Stat(synchronizer.dest)
	if os.IsNotExist(err) {
		fmt.Printf("DEBUG: %s\n", synchronizer.dest)
		err := os.MkdirAll(synchronizer.dest, os.ModePerm)
		checkError(err)
	}
	for _, source := range synchronizer.sources {
		synchronizer.source = source

		// Inicia todos os workers
		synchronizer.wg.Add(synchronizer.NWorkers)
		for i := 1; i <= synchronizer.NWorkers; i++ {
			worker := Worker{synchronizer: synchronizer, id: i}
			worker.init()
			go worker.run()
		}
		synchronizer.updateRecursively(synchronizer.source)
		// Coloca mensagens de finalização para que todos os workers finalizem
		for i := 0; i < synchronizer.NWorkers; i++ {
			synchronizer.jobs <- JobConfig{relPath: "wfinish"}

		}
		synchronizer.wg.Wait() // Aguarda todos os workers terminarem
	}

	synchronizer.results <- ResultData{action: "finish"}

}

// Percorre todas as pastas e diretórios e faz as atualizações necessárias
func (synchronizer *Synchronizer) updateRecursively(path string) {

	if synchronizer.purge {
		relPath, _ := filepath.Rel(synchronizer.source, path)
		destAbsolutePath := filepath.Join(synchronizer.dest, relPath)
		synchronizer.purgeItems(destAbsolutePath)
	}

	items, _ := ioutil.ReadDir(path)
	for _, item := range items {
		sourceAbsolutePath := filepath.Join(path, item.Name())
		relPath, _ := filepath.Rel(synchronizer.source, sourceAbsolutePath)
		if item.IsDir() {
			destAbsolutePath := filepath.Join(synchronizer.dest, relPath)

			//Se não existir a pasta no destino criar
			_, err := os.Stat(destAbsolutePath)
			if os.IsNotExist(err) {
				if synchronizer.verbose {
					fmt.Printf("Nova pasta %s\n", destAbsolutePath)
				}

				err := os.MkdirAll(destAbsolutePath, os.ModePerm)
				checkError(err)
			}
			if synchronizer.purge {
				synchronizer.purgeItems(destAbsolutePath)
			}

			synchronizer.updateRecursively(sourceAbsolutePath)

		} else if synchronizer.autoFind == false || item.Name() != ".sinf_mark.json" {
			var jobConfig JobConfig
			jobConfig.relPath = relPath
			sourceInfo, err := os.Stat(sourceAbsolutePath)
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Erro ao copiar o arquivo %s. Arquivo nao encontrado.\n", sourceAbsolutePath)
				continue
			} else {
				checkError(err)
			}

			size := sourceInfo.Size()
			jobConfig.acknowledge = size > synchronizer.threshold

			synchronizer.jobs <- jobConfig

			// Aguardar copia de arquivo grande
			if jobConfig.acknowledge {
				if size > synchronizer.threshold {
					fmt.Printf("Aguardando arquivo grande: %s %s\n", relPath, humanize.Bytes(uint64(size)))
				}
				<-synchronizer.acknowledgeDone
			}
		}
	}

}

// Copia arquivo mantendo data de modificação e criação inalteradas
// Utiliza dois métodos, caso o arquivo seja pequeno, utiliza a função io.Copy e o progresso só será atualizado no final da cópia.
// No caso de arquivos grandes a cópia será feita bufferizada e por partes informando o gerenciador de progresso a cada pedaço do arquivo copiado.
// A finalidade disso é para que não fique muito tempo sem prints de progresso no console e o usuário fique pensando que o processo travou
// Como o gerenciador de progresso vai sendo informado por partes pode ser que a copia não chegue até o final sem erros.
// Caso houver erro é necessário estornar a quantidade de bytes contabilizadas como já copiados.
// Por isso existe a variável chargeBack que armazena a quantidade de bytes que devem ser estornados em caso de erro
func (synchronizer *Synchronizer) copyFile(src string, dst string, resultData *ResultData) error {

	var err error = nil
	var chargeBack int64

OUTER:
	for i := 0; i < synchronizer.retries; i++ {
		sourceInfo, err := os.Stat(src)
		if os.IsNotExist(err) {
			return err
		}
		checkError(err)

		//Em caso de erro estornar o que já foi informado para o gerenciador de progresso
		if chargeBack != 0 {
			synchronizer.results <- ResultData{action: resultData.action, size: chargeBack, n: 0}
		}

		chargeBack = 0
		resultData.size = 0

		in, err := os.Open(src)
		if err != nil {
			time.Sleep(synchronizer.wait)
			continue
		}
		defer in.Close()

		out, err := os.Create(dst)
		if err != nil {
			time.Sleep(synchronizer.wait)
			continue
		}
		defer out.Close()

		if sourceInfo.Size() > synchronizer.thresholdChunk {
			buf := make([]byte, synchronizer.bufferSize)
			for {
				n, err := in.Read(buf)
				if err != nil && err != io.EOF {
					time.Sleep(synchronizer.wait)
					continue
				}
				if n == 0 {
					break
				}

				if _, err := out.Write(buf[:n]); err != nil {
					time.Sleep(synchronizer.wait)
					continue OUTER
				}
				synchronizer.results <- ResultData{action: resultData.action, n: 0, size: int64(n)}
				chargeBack -= int64(n)
			}
		} else {
			_, err = io.Copy(out, in)
			if err != nil {
				time.Sleep(synchronizer.wait)
				continue
			}
			resultData.size = sourceInfo.Size()
		}

		err = out.Close()
		if err != nil {
			time.Sleep(synchronizer.wait)
			continue
		}

		copyTimes(src, dst)

		// fmt.Printf("Tempos %T %T\n", t.ChangeTime(), t.ModTime())
		// err2 := os.Chtimes(dst, cTime, mTime)
		// if err2 != nil {
		// 	fmt.Fprintf(os.Stderr, "Não foi possível mudar os carimbos de hora do arquivo \"%s\"\n", dst)
		// 	// time.Sleep(synchronizer.wait)
		// 	// continue
		// }

		//Informa gerenciador de progresso que um arquivo terminou sua copia
		resultData.n = 1
		synchronizer.results <- *resultData
		break
	}

	return err
}

//Verifica se existe algum item na pasta de destino que não existe na pasta da fonte e deleta todos
func (synchronizer *Synchronizer) purgeItems(path string) {
	items, _ := ioutil.ReadDir(path)
	for _, item := range items {
		destAbsolutePath := filepath.Join(path, item.Name())
		if synchronizer.autoFind == true && item.Name() == ".sinf_mark.json" {
			continue
		}
		relPath, _ := filepath.Rel(synchronizer.dest, destAbsolutePath)

		//Se não existir a pasta nas fontes deletar
		exists := false
		for _, source := range synchronizer.sources {
			sourceAbsolutePath := filepath.Join(source, relPath)
			_, err := os.Stat(sourceAbsolutePath)
			if !os.IsNotExist(err) {
				exists = true
				break
			}
		}
		if !exists {

			if synchronizer.verbose {
				fmt.Printf("Deletando item %s\n", destAbsolutePath)
			}
			err := os.RemoveAll(destAbsolutePath)
			checkError(err)
			synchronizer.results <- ResultData{action: "delete", size: 0, n: 0}
		}
	}
}

func (synchronizer *Synchronizer) sendIfLogging(data LogInfo) {
	if synchronizer.logging {
		synchronizer.logger.receiver <- data //sends data LogInfo into logger channel
		//fmt.Println(data.logPath + " <- logging path ")
	} else {
		//fmt.Println("Not logging")
	}

}

//Compara arquivo na fonte com destino se for diferente ou não existir copia o novo
func (synchronizer *Synchronizer) copyOrReplaceFile(relPath string) {

	var resultData ResultData
	sourcePath := filepath.Join(synchronizer.source, relPath)
	destPath := filepath.Join(synchronizer.dest, relPath)
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Não foi possível copiar o arquivo \"%s\"\n", sourcePath)
		return
	}
	desttInfo, err := os.Stat(destPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Não foi possível copiar o arquivo \"%s\"\n", sourcePath)
		return
	}
	// Arquivo novo
	if os.IsNotExist(err) {
		if synchronizer.verbose {
			fmt.Printf("Novo arquivo \"%s\"\n", relPath)
		}
		resultData.action = "new"
		err := synchronizer.copyFile(sourcePath, destPath, &resultData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Não foi possível copiar o arquivo \"%s\"\n", sourcePath)
			var log LogInfo
			log.logPath = destPath
			log.logAction = "Failed newfile,"
			synchronizer.sendIfLogging(log)
		} else {
			var log LogInfo
			log.logPath = destPath
			log.logAction = "New,"
			synchronizer.sendIfLogging(log)
		}
		// checkError(err)
		//Arquivo modificado
	} else if (sourceInfo.Size() != desttInfo.Size()) || (sourceInfo.ModTime() != desttInfo.ModTime()) {
		if synchronizer.verbose {
			fmt.Printf("Arquivo alterado \"%s\"\n", relPath)
		}
		resultData.action = "update"
		err := synchronizer.copyFile(sourcePath, destPath, &resultData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Não foi possível copiar o arquivo \"%s\"\n", sourcePath)
			var log LogInfo
			log.logPath = destPath
			log.logAction = "Failed update,"
			synchronizer.sendIfLogging(log)
		} else {
			var log LogInfo
			log.logPath = destPath
			log.logAction = "Updated,"
			synchronizer.sendIfLogging(log)
		}
		// checkError(err)
		// Arquivo igual
	} else {
		resultData.action = "equal"
		resultData.n = 1
		resultData.size = sourceInfo.Size()
		synchronizer.results <- resultData
		var log LogInfo
		log.logPath = destPath
		log.logAction = "Up to date,"
		synchronizer.sendIfLogging(log)
	}

}

func (synchronizer *Synchronizer) foldersOK() {
	for _, source := range synchronizer.sources {
		_, err := os.Stat(source)
		if os.IsNotExist(err) {
			fmt.Printf("O diretório \"%s\" não existe.", source)
			os.Exit(1)
		}
		checkError(err)
	}
}
