package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/djherbis/times.v1"
)

type Files struct {
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
}

func (files *Files) countFiles() {
	files.countFilesRecursively(files.source)

}

//Conta arquivos na pasta de forma recursiva
func (files *Files) countFilesRecursively(path string) {

	items, _ := ioutil.ReadDir(path)
	for _, item := range items {
		if item.IsDir() {
			files.countFilesRecursively(filepath.Join(path, item.Name()))
		} else {
			files.totalNumber += 1
			files.totalSize += item.Size()
		}

	}
}

//Copia arquivo mantendo data de modificação e criação inalterado
func (files *Files) copyFile(src, dst string) error {

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
func (files *Files) copyOrReplaceFile(relPath string) {

	sourcePath := filepath.Join(files.source, relPath)
	destPath := filepath.Join(files.dest, relPath)
	sourceInfo, err := os.Stat(sourcePath)
	checkError(err)
	desttInfo, err := os.Stat(destPath)
	if os.IsNotExist(err) {
		if files.verbose {
			fmt.Printf("Novo arquivo \"%s\"\n", relPath)
		}
		files.copyFile(sourcePath, destPath)
	} else {

		if (sourceInfo.Size() != desttInfo.Size()) || (sourceInfo.ModTime() != desttInfo.ModTime()) {
			if files.verbose {
				fmt.Printf("Arquivo alterado \"%s\"\n", relPath)
			}
			files.copyFile(sourcePath, destPath)
		}
	}
	files.currentNumber = files.currentNumber + 1
	files.currentSize += sourceInfo.Size()
	files.calculateProgress()
}

//Verifica se existe algum item na pasta de destino que não existe na pasta da fonte e deleta todos
func (files *Files) purgeItems(path string) {
	items, _ := ioutil.ReadDir(path)
	for _, item := range items {
		destAbsolutePath := filepath.Join(path, item.Name())
		relPath, _ := filepath.Rel(files.dest, destAbsolutePath)
		sourceAbsolutePath := filepath.Join(files.source, relPath)

		//Se não existir a pasta na fonte deletar
		_, err := os.Stat(sourceAbsolutePath)
		if os.IsNotExist(err) {

			if files.verbose {
				fmt.Printf("Deletando item %s\n", destAbsolutePath)
			}
			err = os.RemoveAll(destAbsolutePath)
			checkError(err)

		}

	}
}

// Calcula o progresso
func (files *Files) calculateProgress() {
	files.progressNumber = 100 * float64(files.currentNumber) / float64(files.totalNumber)
	files.progressSize = 100 * float64(files.currentSize) / float64(files.totalSize)
	files.elapsed = time.Since(files.startTime)
	files.speed = float64(files.currentSize) / float64(files.elapsed.Seconds())
	files.remainingTime = int64(float64(files.totalSize-files.currentSize) / files.speed)
	fmt.Printf("Progresso quantidade: %0.2f%%, Progresso bytes: %0.2f%%\n", files.progressNumber, files.progressSize)
}

func (files *Files) update() {
	files.purgeItems(files.dest)
	files.updateRecursively(files.source)

}

// Percorre todas as pastas e diretórios e faz as atualizações necessárias
func (files *Files) updateRecursively(path string) {
	items, _ := ioutil.ReadDir(path)
	for _, item := range items {
		sourceAbsolutePath := filepath.Join(path, item.Name())
		relPath, _ := filepath.Rel(files.source, sourceAbsolutePath)
		if item.IsDir() {
			destAbsolutePath := filepath.Join(files.dest, relPath)

			//Se não existir a pasta no destino criar
			_, err := os.Stat(destAbsolutePath)
			if os.IsNotExist(err) {
				if files.verbose {
					fmt.Printf("Nova pasta %s\n", destAbsolutePath)
				}

				err := os.MkdirAll(destAbsolutePath, os.ModePerm)
				checkError(err)
			}
			files.purgeItems(destAbsolutePath)
			files.updateRecursively(sourceAbsolutePath)

		} else {
			files.copyOrReplaceFile(relPath)
		}
	}

}

func main() {
	files := Files{
		totalSize:     0,
		totalNumber:   0,
		currentNumber: 0,
		currentSize:   0,
		source:        "D:\\teste_report\\C3",
		dest:          "D:\\teste_report\\c1_copia_deletar",
		verbose:       true}
	files.startTime = time.Now()
	// files.calculateElapsed()
	fmt.Println("Contando arquivos...")
	files.countFiles()
	fmt.Printf("Quantidade total: %d, Tamanho total: %d \n", files.totalNumber, files.totalSize)
	fmt.Println("Sincronizando...")
	files.update()

	log.Printf("It took %s", files.elapsed)
	fmt.Println(files.totalSize, files.totalNumber)
}
