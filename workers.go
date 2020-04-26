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

//Copia arquivo mantendo data de modificação e criação inalteradas
func copyFile(config *Config, src string, dst string, results chan<- ResultData, resultData *ResultData) error {

	var err error = nil
	var chargeBack int64

OUTER:
	for i := 0; i < config.retries; i++ {
		sourceInfo, err := os.Stat(src)
		if os.IsNotExist(err) {
			return err
		}
		checkError(err)

		// Caso o arquivo seja grande anotar para copiar depois
		// if !config.copyLargeFiles && sourceInfo.Size() > config.threshold {
		// 	fmt.Printf("DEBUG LARGE FILE: %s\n", src)
		// 	results <- ResultData{action: "large_file", path: src}
		// 	return err
		// }

		//Em caso de erro estornar o que já foi informado para o gerenciador de progresso
		if chargeBack != 0 {
			results <- ResultData{action: "correction", size: chargeBack, n: 0}
		}

		chargeBack = 0
		resultData.size = 0

		in, err := os.Open(src)
		if err != nil {
			time.Sleep(config.wait)
			continue
		}
		defer in.Close()

		out, err := os.Create(dst)
		if err != nil {
			time.Sleep(config.wait)
			continue
		}
		defer out.Close()

		//Caso o arquivo seja grande copia por partes e atualza progresso por partes
		if sourceInfo.Size() > 1000000000 {

			// fmt.Printf("Copiando arquivo grande \"%s\", tamanho: %s\n", src, humanize.Bytes(uint64(sourceInfo.Size())))
			buf := make([]byte, config.bufferSize)
			for {
				n, err := in.Read(buf)
				if err != nil && err != io.EOF {
					time.Sleep(config.wait)
					continue
				}
				if n == 0 {
					break
				}

				if _, err := out.Write(buf[:n]); err != nil {
					time.Sleep(config.wait)
					continue OUTER
				}
				results <- ResultData{action: "partial_copy", n: 0, size: int64(n)}
				chargeBack -= int64(n)
			}
		} else {
			_, err = io.Copy(out, in)
			if err != nil {
				time.Sleep(config.wait)
				continue
			}
			resultData.size = sourceInfo.Size()
		}

		err = out.Close()
		if err != nil {
			time.Sleep(config.wait)
			continue
		}
		t, err := times.Stat(src)
		if err != nil {
			time.Sleep(config.wait)
			continue
		}
		err = os.Chtimes(dst, t.ChangeTime(), t.ModTime())
		if err != nil {
			time.Sleep(config.wait)
			continue
		}

		//Informa gerenciador de progresso que um arquivo terminou sua copia
		resultData.n = 1
		results <- *resultData
		break
	}

	return err
}

//Verifica se existe algum item na pasta de destino que não existe na pasta da fonte e deleta todos
func purgeItems(config *Config, path string, results chan<- ResultData) {
	items, _ := ioutil.ReadDir(path)
	for _, item := range items {
		destAbsolutePath := filepath.Join(path, item.Name())
		relPath, _ := filepath.Rel(config.dest, destAbsolutePath)
		sourceAbsolutePath := filepath.Join(config.source, relPath)

		//Se não existir a pasta na fonte deletar
		_, err := os.Stat(sourceAbsolutePath)
		if os.IsNotExist(err) {

			if config.verbose {
				fmt.Printf("Deletando item %s\n", destAbsolutePath)
			}
			err = os.RemoveAll(destAbsolutePath)
			checkError(err)
			results <- ResultData{action: "delete", size: 0, n: 0}
		}
	}
}

//Compara arquivo na fonte com destino se for diferente ou não existir copia o novo
func copyOrReplaceFile(config *Config, relPath string, results chan<- ResultData) {

	var resultData ResultData
	sourcePath := filepath.Join(config.source, relPath)
	destPath := filepath.Join(config.dest, relPath)
	sourceInfo, err := os.Stat(sourcePath)
	checkError(err)
	desttInfo, err := os.Stat(destPath)
	if os.IsNotExist(err) { // Arquivo novo
		if config.verbose {
			fmt.Printf("Novo arquivo \"%s\"\n", relPath)
		}
		resultData.action = "new"
		err := copyFile(config, sourcePath, destPath, results, &resultData)
		checkError(err)

	} else if (sourceInfo.Size() != desttInfo.Size()) || (sourceInfo.ModTime() != desttInfo.ModTime()) { //Arquivo modificado
		if config.verbose {
			fmt.Printf("Arquivo alterado \"%s\"\n", relPath)
		}
		resultData.action = "update"
		err := copyFile(config, sourcePath, destPath, results, &resultData)
		checkError(err)

	} else { // Arquivo igual
		resultData.action = "equal"
		resultData.n = 1
		resultData.size = sourceInfo.Size()
		results <- resultData
	}

}

type WorkerConfig struct {
	config      *Config
	relPath     string
	acknowledge bool // Flag para dizer ao work que ele deve informar que terminou a tarefa
}

//Compara arquivo na fonte com destino se for diferente ou não existir copia o novo
func fileWorker(jobs <-chan WorkerConfig, results chan<- ResultData, wg *sync.WaitGroup, id int, acknowledgeDone chan bool) {

	defer func() {
		wg.Done()
	}()
	for j := range jobs {
		if j.relPath != "wfinish" {
			copyOrReplaceFile(j.config, j.relPath, results)
			if j.acknowledge {
				acknowledgeDone <- true
			}
		} else {
			return
		}
	}

}

func update(config *Config, jobs chan WorkerConfig, results chan ResultData) {

	//Cria a pasta de destino caso ela não exista
	_, err := os.Stat(config.dest)
	if os.IsNotExist(err) {
		err := os.MkdirAll(config.dest, os.ModePerm)
		checkError(err)
	}

	var wg sync.WaitGroup
	acknowledgeDone := make(chan bool)

	// Inicia todos os workers
	wg.Add(config.NWorkers)
	for i := 1; i <= config.NWorkers; i++ {
		go fileWorker(jobs, results, &wg, i, acknowledgeDone)
	}

	updateRecursively(config, config.source, jobs, results, acknowledgeDone)

	// Coloca mensagens de finalização para que todos os workers finalizem
	for i := 0; i < config.NWorkers; i++ {
		jobs <- WorkerConfig{relPath: "wfinish"}

	}
	wg.Wait() // Aguarda todos os workers terminarem

	// if config.verbose {
	// 	fmt.Println("Iniciando copia de arquivos grandes.")
	// }
	// config.copyLargeFiles = true
	// for _, relPath := range *largeFiles {

	// 	copyOrReplaceFile(config, relPath, results)
	// }

	results <- ResultData{action: "finish"}

}

// Percorre todas as pastas e diretórios e faz as atualizações necessárias
func updateRecursively(config *Config, path string, jobs chan<- WorkerConfig, results chan<- ResultData, acknowledgeDone <-chan bool) {

	if config.purge {
		relPath, _ := filepath.Rel(config.source, path)
		destAbsolutePath := filepath.Join(config.dest, relPath)
		purgeItems(config, destAbsolutePath, results)
	}

	items, _ := ioutil.ReadDir(path)
	for _, item := range items {
		sourceAbsolutePath := filepath.Join(path, item.Name())
		relPath, _ := filepath.Rel(config.source, sourceAbsolutePath)
		if item.IsDir() {
			destAbsolutePath := filepath.Join(config.dest, relPath)

			//Se não existir a pasta no destino criar
			_, err := os.Stat(destAbsolutePath)
			if os.IsNotExist(err) {
				if config.verbose {
					fmt.Printf("Nova pasta %s\n", destAbsolutePath)
				}

				err := os.MkdirAll(destAbsolutePath, os.ModePerm)
				checkError(err)
			}
			if config.purge {
				purgeItems(config, destAbsolutePath, results)
			}

			updateRecursively(config, sourceAbsolutePath, jobs, results, acknowledgeDone)

		} else {
			var workerConfig WorkerConfig
			workerConfig.config = config
			workerConfig.relPath = relPath
			sourceInfo, err := os.Stat(sourceAbsolutePath)
			checkError(err)
			workerConfig.acknowledge = sourceInfo.Size() > config.threshold

			jobs <- workerConfig

			// Aguardar copia de arquivo grande
			if workerConfig.acknowledge {
				fmt.Println("DEBUG: Aguardando terminar copia de arquivo grande")
				<-acknowledgeDone
			}
		}
	}

}

type ResultData struct {
	size   int64
	n      int64
	action string
	path   string
}

// Worker responsável por calcular o progresso e printar no console.
func progressWork(config *Config, progress *Progress, results <-chan ResultData, finished chan<- bool) {

	for resultData := range results {

		if resultData.action == "finish" {
			break
		}
		switch resultData.action {
		case "new":
			progress.currentNumber += resultData.n
			progress.currentSize += resultData.size
			progress.newFiles++
		case "update":
			progress.currentNumber += resultData.n
			progress.currentSize += resultData.size
			progress.updateFiles++
		case "delete":
			progress.currentNumber += resultData.n
			progress.currentSize += resultData.size
			progress.deletedItems++
		case "equal":
			progress.currentNumber += resultData.n
			progress.currentSize += resultData.size
			progress.equalFiles++
		case "correction":
			progress.currentNumber += resultData.n
			progress.currentSize += resultData.size
		}

		progress.calculateProgress()
	}
	// progress.calculateProgress()
	finished <- true
}
