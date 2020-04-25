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
	"gopkg.in/djherbis/times.v1"
)

//Copia arquivo mantendo data de modificação e criação inalterado
func copyFile(config *Config, src string, dst string, results chan<- ResultData, resultData *ResultData) error {

	var err error = nil
	var chargeBack int64
	// var pending int64
	// resultData := ResultData{size:}

OUTER:
	for i := 0; i < config.retries; i++ {

		//Em caso de erro estornar o que já foi informado para o gerenciador de progresso
		if chargeBack != 0 {
			results <- ResultData{action: "correction", size: chargeBack, n: 0}
		}

		chargeBack = 0
		resultData.size = 0

		sourceInfo, err := os.Stat(src)
		if err != nil {
			time.Sleep(config.wait)
			continue
		}

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

			fmt.Printf("Copiando arquivo grande \"%s\", tamanho: %s\n", src, humanize.Bytes(uint64(sourceInfo.Size())))
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

//Copia arquivo mantendo data de modificação e criação inalterado
// func copyFile(config *Config, src string, dst string) error {
// 	var err error = nil
// 	// Tenta copiar arquivo várias vezes
// 	for i := 0; i < config.retries; i++ {
// 		sourceInfo, err := os.Stat(src)
// 		if err != nil {
// 			time.Sleep(config.wait)
// 			continue
// 		}
// 		if sourceInfo.Size() > 5000000000 {
// 			fmt.Printf("Copiando arquivo\"%s\", tamanho: %s\n", src, humanize.Bytes(uint64(sourceInfo.Size())))
// 		}

// 		in, err := os.Open(src)
// 		if err != nil {
// 			time.Sleep(config.wait)
// 			continue
// 		}
// 		defer in.Close()

// 		out, err := os.Create(dst)
// 		if err != nil {
// 			time.Sleep(config.wait)
// 			continue
// 		}
// 		defer out.Close()

// 		buf := make([]byte, 512)
// 		for {
// 			n, err := in.Read(buf)
// 			if err != nil && err != io.EOF {
// 				time.Sleep(config.wait)
// 				continue
// 			}
// 			if n == 0 {
// 				break
// 			}

// 			if _, err := out.Write(buf[:n]); err != nil {
// 				time.Sleep(config.wait)
// 				continue
// 			}
// 		}

// 		err = out.Close()
// 		if err != nil {
// 			time.Sleep(config.wait)
// 			continue
// 		}
// 		t, err := times.Stat(src)
// 		if err != nil {
// 			time.Sleep(config.wait)
// 			continue
// 		}
// 		err = os.Chtimes(dst, t.ChangeTime(), t.ModTime())
// 		if err != nil {
// 			time.Sleep(config.wait)
// 			continue
// 		}
// 		break
// 	}
// 	return err

// }

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
		err := copyFile(config, sourcePath, destPath, results, &resultData)
		checkError(err)

		resultData.action = "new"
	} else if (sourceInfo.Size() != desttInfo.Size()) || (sourceInfo.ModTime() != desttInfo.ModTime()) { //Arquivo modificado
		if config.verbose {
			fmt.Printf("Arquivo alterado \"%s\"\n", relPath)
		}
		copyFile(config, sourcePath, destPath, results, &resultData)
		resultData.action = "update"
	} else { // Arquivo igual
		resultData.action = "equal"
		resultData.n = 1
		resultData.size = sourceInfo.Size()
		results <- resultData
	}

}

type WorkerConfig struct {
	config  *Config
	relPath string
}

//Compara arquivo na fonte com destino se for diferente ou não existir copia o novo
func fileWorker(jobs <-chan WorkerConfig, results chan<- ResultData, wg *sync.WaitGroup, id int) {

	defer func() {
		wg.Done()
	}()
	for j := range jobs {
		if j.relPath != "wfinish" {
			copyOrReplaceFile(j.config, j.relPath, results)
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

	wg.Add(config.NWorkers)
	for i := 1; i <= config.NWorkers; i++ {
		go fileWorker(jobs, results, &wg, i)
	}

	updateRecursively(config, config.source, jobs, results)
	for i := 0; i < config.NWorkers; i++ {
		jobs <- WorkerConfig{relPath: "wfinish"}

	}
	wg.Wait()
	results <- ResultData{action: "finish"}
}

// Percorre todas as pastas e diretórios e faz as atualizações necessárias
func updateRecursively(config *Config, path string, jobs chan<- WorkerConfig, results chan<- ResultData) {

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

			updateRecursively(config, sourceAbsolutePath, jobs, results)

		} else {
			var workerConfig WorkerConfig
			workerConfig.config = config
			workerConfig.relPath = relPath
			jobs <- workerConfig
			// copyOrReplaceFile(config, relPath)
		}
	}

}

type ResultData struct {
	size   int64
	n      int64
	action string
}

func progressWork(config *Config, progress *Progress, results <-chan ResultData, finished chan<- bool) {

	for resultData := range results {

		if resultData.action == "finish" {
			break
		}
		progress.currentNumber += resultData.n
		progress.currentSize += resultData.size
		switch resultData.action {
		case "new":
			progress.newFiles++
		case "update":
			progress.updateFiles++
		case "delete":
			progress.deletedItems++
		case "equal":
			progress.equalFiles++
		}

		progress.calculateProgress()
	}
	// progress.calculateProgress()
	finished <- true
}
