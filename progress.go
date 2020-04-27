package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
)

// Progress armazena informações sobre o progresso
type Progress struct {
	totalSize         int64             // Quantidade total de bytes
	totalNumber       int64             // Quantidade total de arquivos
	currentSize       int64             // Quantidade de bytes analisados
	currentNumber     int64             // Quantidade de arquivos analisados
	totalToCopySize   int64             // Quantidade total para copiar em bytes desconsiderando os que já estão iguais
	totalToCopyNumber int64             // Quantidade total para copiar em número desconsiderando os que já estão iguais
	sizeAnalyzed      int64             // Quantidade de bytes já atualizados
	numberAnalyzed    int64             // Quantidade de arquivos já atualizados
	sizeCopied        int64             // Quantidade de bytes copiados
	numberCopied      int64             // Quantidade de arquivos copiados
	startTime         time.Time         // Horário de inicio do processamento
	eta               time.Time         // Horario de termino
	eta2              time.Time         // Horário de término utilizando velocidade estimada
	elapsed           time.Duration     // Tempo decorrido desde o início
	avgSpeed          float64           // Velocidade média em bytes por nanosegundo
	speed             float64           // Velocidade instantânea em bytes por nanosegundo
	remainingTime     time.Duration     // Tempo que falta
	remainingTime2    time.Duration     // Tempo que falta com a velocidade final estimada
	progressSize      float64           // Progresso percentual em quantidade de bytes
	progressNumber    float64           // Progresso percentual em quantidade de arquivos
	newFiles          int64             // Quantidade de novos arquivos copiados
	updatedFiles      int64             // Quantidae de arquivos modificados
	deletedItems      int64             // Quantidade de itens deletados, pastas e arquivos
	equalFiles        int64             // Quantidate de arquivos iguais
	results           <-chan ResultData // Canal utilizado para passar os resultado das tarefas para o gerenciador de progresso
	finished          chan bool         // Canal utilizado para informar o sincronizador que o trabalho foi finalizado
	synchronizer      *Synchronizer     // Ponteiro a estrutura do sincronizador
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
	progress.updatedFiles = 0
	progress.deletedItems = 0
	progress.totalSize = 0
	progress.currentNumber = 0
	progress.currentSize = 0
	progress.sizeCopied = 0
	progress.numberCopied = 0
	progress.results = progress.synchronizer.results
	progress.finished = progress.synchronizer.finished
}

// // Calcula o progresso e imprime no console
// func (progress *Progress) calculateProgress() {
// 	progress.progressSize = float64(progress.currentSize) / float64(progress.totalSize)
// 	progress.progressNumber = float64(progress.currentNumber) / float64(progress.totalNumber)
// 	progress.elapsed = time.Since(progress.startTime)
// 	progress.avgSpeed = float64(progress.currentSize) / float64(progress.elapsed.Nanoseconds()) //bytes por nanosegundo
// 	alfa := (float64(progress.totalSize-progress.currentSize) / (float64(progress.totalNumber - progress.currentNumber))) / (float64(progress.currentSize) / float64(progress.currentNumber))
// 	alfa = 0.6*alfa + 0.4
// 	estimatedSpeed := alfa * progress.avgSpeed
// 	progress.remainingTime = int64(float64(progress.totalSize-progress.currentSize) / estimatedSpeed)
// 	remainingTimeStr := fmtDuration(time.Duration(progress.remainingTime))
// 	speed := progress.avgSpeed * 1000000000 * 60      //bytes por minuto
// 	estimatedSpeed = estimatedSpeed * 1000000000 * 60 //bytes por minuto
// 	fmt.Printf("Velocidade média: %s/min    Velocidade média final estimada: %s/min   %d de %d    %0.2f%%    %s\n", humanize.Bytes(uint64(speed)), humanize.Bytes(uint64(estimatedSpeed)), progress.currentNumber, progress.totalNumber, 100*progress.progressSize, remainingTimeStr)

// }

func (progress *Progress) calculateProgress(action string) {
	var speed float64
	var estimatedSpeed float64
	var remainingTimeStr2 string
	var remainingTimeStr string
	var etaStr string
	var eta2Str string
	progress.progressSize = float64(progress.sizeCopied) / float64(progress.totalToCopySize)
	progress.progressNumber = float64(progress.numberCopied) / float64(progress.totalToCopyNumber)
	progress.elapsed = time.Since(progress.startTime)
	progress.avgSpeed = float64(progress.sizeCopied) / float64(progress.elapsed.Nanoseconds()) //bytes por nanosegundo
	if progress.progressSize == 0 {
		estimatedSpeed = 0
		speed = 0
		estimatedSpeed = 0
		remainingTimeStr = "-"
		remainingTimeStr2 = "-"
		etaStr = "-"
		eta2Str = "-"
	} else {
		alfa := (progress.progressNumber / progress.progressSize)
		alfa = 0.6*alfa + 0.4 // Correção do alfa para que varie de forma menos agressiva
		estimatedSpeed = alfa * progress.avgSpeed
		progress.remainingTime = time.Duration(int64(float64(progress.totalToCopySize-progress.sizeCopied) / progress.avgSpeed))
		progress.remainingTime2 = time.Duration(int64(float64(progress.totalToCopySize-progress.sizeCopied) / estimatedSpeed))
		progress.eta = time.Now().Local().Add(progress.remainingTime)
		progress.eta2 = time.Now().Local().Add(progress.remainingTime2)
		speed = progress.avgSpeed * 1000000000 * 60
		remainingTimeStr = fmtDuration(progress.remainingTime)
		remainingTimeStr2 = fmtDuration(progress.remainingTime2)
		etaStr = fmtTime(progress.eta)
		eta2Str = fmtTime(progress.eta2)
	}

	fmt.Printf("%-8v  %s/min  %d/%d    %0.2f%%      %s  %s       %s  %s \n", action, humanize.Bytes(uint64(speed)), progress.currentNumber, progress.totalNumber, 100*progress.progressSize, remainingTimeStr, etaStr, remainingTimeStr2, eta2Str)

}

// Conta os arquivos e calcula o tamanho total em bytes
func (progress *Progress) countFiles() {
	progress.totalNumber = 0
	progress.totalSize = 0
	progress.countFilesRecursively(progress.synchronizer.source)
	progress.totalToCopyNumber = progress.totalNumber
	progress.totalToCopySize = progress.totalSize
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
			progress.sizeCopied += resultData.size
			progress.numberCopied += resultData.n
			progress.newFiles++
		case "update":
			progress.currentNumber += resultData.n
			progress.currentSize += resultData.size
			progress.sizeCopied += resultData.size
			progress.numberCopied += resultData.n
			progress.updatedFiles++
		case "delete":
			progress.currentNumber += resultData.n
			progress.currentSize += resultData.size
			progress.deletedItems++
		case "equal":
			progress.currentNumber += resultData.n
			progress.currentSize += resultData.size
			progress.totalToCopySize -= resultData.size
			progress.totalToCopyNumber -= resultData.n
			progress.equalFiles++
		}

		progress.calculateProgress(resultData.action)
	}

	progress.finished <- true
}
