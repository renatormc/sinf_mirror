package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/djherbis/times.v1"
)

//Copia arquivo mantendo data de modificação e criação inalterado
func copyFile(src, dst string) error {

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	err = out.Close()
	if err != nil {
		return err
	}
	t, err := times.Stat(src)
	if err != nil {
		return err
	}
	err = os.Chtimes(dst, t.ChangeTime(), t.ModTime())
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
			results <- ResultData{action: "delete", size: 0}

		}

	}
}

//Compara arquivo na fonte com destino se for diferente ou não existir copia o novo
func copyOrReplaceFile(config *Config, relPath string, results chan<- ResultData) {

	var result ResultData
	defer func() {
		results <- result
	}()
	sourcePath := filepath.Join(config.source, relPath)
	destPath := filepath.Join(config.dest, relPath)
	sourceInfo, err := os.Stat(sourcePath)
	checkError(err)
	desttInfo, err := os.Stat(destPath)
	if os.IsNotExist(err) {
		if config.verbose {
			fmt.Printf("Novo arquivo \"%s\"\n", relPath)
		}
		copyFile(sourcePath, destPath)
		result.action = "new"
	} else {

		if (sourceInfo.Size() != desttInfo.Size()) || (sourceInfo.ModTime() != desttInfo.ModTime()) {
			if config.verbose {
				fmt.Printf("Arquivo alterado \"%s\"\n", relPath)
			}
			copyFile(sourcePath, destPath)
			result.action = "update"
		}
	}
	result.size = sourceInfo.Size()

}

type WorkerConfig struct {
	config  *Config
	relPath string
}

//Compara arquivo na fonte com destino se for diferente ou não existir copia o novo
func fileWorker(jobs <-chan WorkerConfig, results chan<- ResultData, wg *sync.WaitGroup, id int) {

	defer func() {
		fmt.Printf("Worker %d finalizou\n", id)
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
	var wg sync.WaitGroup

	wg.Add(config.NWorkers)
	for i := 1; i <= config.NWorkers; i++ {

		go fileWorker(jobs, results, &wg, i)
		fmt.Printf("Iniciando worker %d\n", i)
	}

	updateRecursively(config, config.source, jobs, results)
	for i := 0; i < config.NWorkers; i++ {
		jobs <- WorkerConfig{relPath: "wfinish"}

	}
	wg.Wait()
	fmt.Println("DEBUG: Passou wait")
	results <- ResultData{size: -1}

}

// Percorre todas as pastas e diretórios e faz as atualizações necessárias
func updateRecursively(config *Config, path string, jobs chan<- WorkerConfig, results chan<- ResultData) {
	// defer close(jobs)

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
			purgeItems(config, destAbsolutePath, results)
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
	action string
}

func progressWork(config *Config, progress *Progress, results <-chan ResultData, finished chan<- bool) {

	for resultData := range results {

		if resultData.size == -1 {
			fmt.Println("DEBUG: recebeu -1")
			break
		}
		progress.currentNumber++
		progress.currentSize += resultData.size
		switch resultData.action {
		case "new":
			progress.newFiles++
		case "update":
			progress.updateFiles++
		case "delete":
			progress.deletedItems++
		}

		progress.calculateProgress()
	}
	// progress.calculateProgress()
	finished <- true
}
