package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/djherbis/times.v1"
)

// Synchronizer matém as configuraçõse do programa
type Synchronizer struct {
	source          string        //pasta fonte
	dest            string        // pasta destino
	verbose         bool          // Imprimir mensagens extras
	NWorkers        int           // Número de workers
	threshold       int64         // Tamanho em megabytes a partir do qual será copiado sem concorrência
	bufferSize      int64         // Tamanho do buffer utilizado para copiar arquivos grandes
	purge           bool          // Deletar o que existir no destino e não na fonte
	retries         int           // Número de tentativas de copiar arquivo caso ocorra erros
	wait            time.Duration // Tempo para aguardar antes de tentar de novo em segundos
	results         chan ResultData
	jobs            chan JobConfig
	acknowledgeDone chan bool
	finished        chan bool
	wg              sync.WaitGroup
}

func (synchronizer *Synchronizer) init() {
	synchronizer.jobs = make(chan JobConfig)
	synchronizer.results = make(chan ResultData)
	synchronizer.finished = make(chan bool)
}

func (synchronizer *Synchronizer) run() {

	//Cria a pasta de destino caso ela não exista
	_, err := os.Stat(synchronizer.dest)
	if os.IsNotExist(err) {
		err := os.MkdirAll(synchronizer.dest, os.ModePerm)
		checkError(err)
	}

	synchronizer.acknowledgeDone = make(chan bool)

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

		} else {
			var jobConfig JobConfig
			jobConfig.relPath = relPath
			sourceInfo, err := os.Stat(sourceAbsolutePath)
			checkError(err)
			jobConfig.acknowledge = sourceInfo.Size() > synchronizer.threshold

			synchronizer.jobs <- jobConfig

			// Aguardar copia de arquivo grande
			if jobConfig.acknowledge {
				fmt.Println("DEBUG: Aguardando terminar copia de arquivo grande")
				<-synchronizer.acknowledgeDone
			}
		}
	}

}

//Copia arquivo mantendo data de modificação e criação inalteradas
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
			synchronizer.results <- ResultData{action: "correction", size: chargeBack, n: 0}
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

		//Caso o arquivo seja grande copia por partes e atualza progresso por partes
		if sourceInfo.Size() > 1000000000 {

			// fmt.Printf("Copiando arquivo grande \"%s\", tamanho: %s\n", src, humanize.Bytes(uint64(sourceInfo.Size())))
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
				synchronizer.results <- ResultData{action: "partial_copy", n: 0, size: int64(n)}
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
		t, err := times.Stat(src)
		if err != nil {
			time.Sleep(synchronizer.wait)
			continue
		}
		err = os.Chtimes(dst, t.ChangeTime(), t.ModTime())
		if err != nil {
			time.Sleep(synchronizer.wait)
			continue
		}

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
		relPath, _ := filepath.Rel(synchronizer.dest, destAbsolutePath)
		sourceAbsolutePath := filepath.Join(synchronizer.source, relPath)

		//Se não existir a pasta na fonte deletar
		_, err := os.Stat(sourceAbsolutePath)
		if os.IsNotExist(err) {

			if synchronizer.verbose {
				fmt.Printf("Deletando item %s\n", destAbsolutePath)
			}
			err = os.RemoveAll(destAbsolutePath)
			checkError(err)
			synchronizer.results <- ResultData{action: "delete", size: 0, n: 0}
		}
	}
}

//Compara arquivo na fonte com destino se for diferente ou não existir copia o novo
func (synchronizer *Synchronizer) copyOrReplaceFile(relPath string) {

	var resultData ResultData
	sourcePath := filepath.Join(synchronizer.source, relPath)
	destPath := filepath.Join(synchronizer.dest, relPath)
	sourceInfo, err := os.Stat(sourcePath)
	checkError(err)
	desttInfo, err := os.Stat(destPath)
	if os.IsNotExist(err) { // Arquivo novo
		if synchronizer.verbose {
			fmt.Printf("Novo arquivo \"%s\"\n", relPath)
		}
		resultData.action = "new"
		err := synchronizer.copyFile(sourcePath, destPath, &resultData)
		checkError(err)

	} else if (sourceInfo.Size() != desttInfo.Size()) || (sourceInfo.ModTime() != desttInfo.ModTime()) { //Arquivo modificado
		if synchronizer.verbose {
			fmt.Printf("Arquivo alterado \"%s\"\n", relPath)
		}
		resultData.action = "update"
		err := synchronizer.copyFile(sourcePath, destPath, &resultData)
		checkError(err)

	} else { // Arquivo igual
		resultData.action = "equal"
		resultData.n = 1
		resultData.size = sourceInfo.Size()
		synchronizer.results <- resultData
	}

}
