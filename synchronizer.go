package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/djherbis/times.v1"
)

type Synchronizer struct {
	totalSize      int64
	totalNumber    int64
	currentSize    int64
	currentNumber  int64
	source         string
	dest           string
	verbose        bool
	startTime      time.Time
	elapsed        time.Duration
	speed          float64
	remainingTime  int64
	progressSize   float64
	progressNumber float64
	appConfig      AppConfig
}

func (synchronizer *Synchronizer) countFiles() {
	synchronizer.countFilesRecursively(synchronizer.source)

}

//Conta arquivos na pasta de forma recursiva
func (synchronizer *Synchronizer) countFilesRecursively(path string) {

	items, _ := ioutil.ReadDir(path)
	for _, item := range items {
		if item.IsDir() {
			synchronizer.countFilesRecursively(filepath.Join(path, item.Name()))
		} else {
			synchronizer.totalNumber += 1
			synchronizer.totalSize += item.Size()
		}

	}
}

//Copia arquivo mantendo data de modificação e criação inalterado
func (synchronizer *Synchronizer) copyFile(src, dst string) error {

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

//Compara arquivo na fonte com destino se for diferente ou não existir copia o novo
func (synchronizer *Synchronizer) copyOrReplaceFile(relPath string) {

	sourcePath := filepath.Join(synchronizer.source, relPath)
	destPath := filepath.Join(synchronizer.dest, relPath)
	sourceInfo, err := os.Stat(sourcePath)
	checkError(err)
	desttInfo, err := os.Stat(destPath)
	if os.IsNotExist(err) {
		if synchronizer.verbose {
			fmt.Printf("Novo arquivo \"%s\"\n", relPath)
		}
		synchronizer.copyFile(sourcePath, destPath)
	} else {

		if (sourceInfo.Size() != desttInfo.Size()) || (sourceInfo.ModTime() != desttInfo.ModTime()) {
			if synchronizer.verbose {
				fmt.Printf("Arquivo alterado \"%s\"\n", relPath)
			}
			synchronizer.copyFile(sourcePath, destPath)
		}
	}
	synchronizer.currentNumber = synchronizer.currentNumber + 1
	synchronizer.currentSize += sourceInfo.Size()
	synchronizer.calculateProgress()
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

		}

	}
}

// Calcula o progresso
func (synchronizer *Synchronizer) calculateProgress() {
	synchronizer.progressNumber = 100 * float64(synchronizer.currentNumber) / float64(synchronizer.totalNumber)
	synchronizer.progressSize = 100 * float64(synchronizer.currentSize) / float64(synchronizer.totalSize)
	synchronizer.elapsed = time.Since(synchronizer.startTime)
	synchronizer.speed = float64(synchronizer.currentSize) / float64(synchronizer.elapsed.Seconds())
	synchronizer.remainingTime = int64(float64(synchronizer.totalSize-synchronizer.currentSize) / synchronizer.speed)
	fmt.Printf("Progresso quantidade: %0.2f%%, Progresso bytes: %0.2f%%\n", synchronizer.progressNumber, synchronizer.progressSize)
}

func (synchronizer *Synchronizer) update() {
	synchronizer.purgeItems(synchronizer.dest)
	synchronizer.updateRecursively(synchronizer.source)

}

// Percorre todas as pastas e diretórios e faz as atualizações necessárias
func (synchronizer *Synchronizer) updateRecursively(path string) {
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
			synchronizer.purgeItems(destAbsolutePath)
			synchronizer.updateRecursively(sourceAbsolutePath)

		} else {
			synchronizer.copyOrReplaceFile(relPath)
		}
	}

}
