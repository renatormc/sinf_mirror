package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hako/durafmt"
)

// Progress armazena informações sobre o progresso
type Progress struct {
	totalSize     int64             // Quantidade total de bytes
	totalNumber   int64             // Quantidade total de arquivos
	currentSize   int64             // Quantidade de bytes já atualizados
	currentNumber int64             // Quantidade de arquivos já atualizados
	startTime     time.Time         // Horário de inicio do processamento
	elapsed       time.Duration     // Tempo decorrido desde o início
	speed         float64           // Velocidade em bytes por nanosegundo
	remainingTime int64             // Tempo que falta
	progressSize  float64           // Progresso percentual em quantidade de bytes
	newFiles      int64             // Quantidade de novos arquivos copiados
	updateFiles   int64             // Quantidae de arquivos modificados
	deletedItems  int64             // Quantidade de itens deletados, pastas e arquivos
	equalFiles    int64             // Quantidate de arquivos iguais
	results       <-chan ResultData // Canal utilizado para passar os resultado das tarefas para o gerenciador de progresso
	finished      chan bool         // Canal utilizado para informar o sincronizador que o trabalho foi finalizado
	synchronizer  *Synchronizer     // Ponteiro a estrutura do sincronizador
}

// ResultData serve para passar dados dos workers para o gerenciador de progresso
type ResultData struct {
	size   int64
	n      int64
	action string
	path   string
}

func (progress *Progress) init() {
	progress.newFiles = 0
	progress.updateFiles = 0
	progress.deletedItems = 0
	progress.totalSize = 0
	progress.results = progress.synchronizer.results
	progress.finished = progress.synchronizer.finished
}

// Calcula o progresso e imprime no console
func (progress *Progress) calculateProgress() {
	progress.progressSize = 100 * float64(progress.currentSize) / float64(progress.totalSize)
	progress.elapsed = time.Since(progress.startTime)
	progress.speed = float64(progress.currentSize) / float64(progress.elapsed.Nanoseconds()) //bytes por nanosegundo
	progress.remainingTime = int64(float64(progress.totalSize-progress.currentSize) / progress.speed)
	remainingTimeStr := durafmt.Parse(time.Duration(progress.remainingTime)).String()
	speed := progress.speed * 1000000000 * 60 //bytes por minuto
	fmt.Printf("%s/min    %d de %d    %0.2f%%    Estimado: %s\n", humanize.Bytes(uint64(speed)), progress.currentNumber, progress.totalNumber, progress.progressSize, remainingTimeStr)

}

// Conta os arquivos e calcula o tamanho total em bytes
func (progress *Progress) countFiles() {
	progress.totalNumber = 0
	progress.totalSize = 0
	progress.countFilesRecursively(progress.synchronizer.source)
}

func (progress *Progress) countFilesRecursively(path string) {

	items, _ := ioutil.ReadDir(path)
	for _, item := range items {
		if item.IsDir() {
			progress.countFilesRecursively(filepath.Join(path, item.Name()))
		} else {
			progress.totalNumber++
			progress.totalSize += item.Size()
		}
	}
}

// Função responsável por receber as mensagens de finalização de cada trabalho.
// Finaliza quando recebe a mensagem contendo no campo action a palavra "finish" então informa o sincronizador que o trabalho todo terminou
func (progress *Progress) run() {
	progress.startTime = time.Now()

	for resultData := range progress.results {

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

	progress.finished <- true
}
