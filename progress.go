package main

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hako/durafmt"
)

// Estrutura que armazena informações sobre o progress
type Progress struct {
	totalSize     int64         // Quantidade total de bytes
	totalNumber   int64         // Quantidade total de arquivos
	currentSize   int64         // Quantidade de bytes já atualizados
	currentNumber int64         // Quantidade de arquivos já atualizados
	startTime     time.Time     // Horário de inicio do processamento
	elapsed       time.Duration // Tempo decorrido desde o início
	speed         float64       // Velocidade em bytes por nanosegundo
	remainingTime int64         // Tempo que falta
	progressSize  float64       // Progresso percentual em quantidade de bytes
	newFiles      int64         // Quantidade de novos arquivos copiados
	updateFiles   int64         // Quantidae de arquivos modificados
	deletedItems  int64         // Quantidade de itens deletados, pastas e arquivos
	equalFiles    int64         // Quantidate de arquivos iguais
}

// Calcula o progresso
func (progress *Progress) calculateProgress() {
	progress.progressSize = 100 * float64(progress.currentSize) / float64(progress.totalSize)
	progress.elapsed = time.Since(progress.startTime)
	progress.speed = float64(progress.currentSize) / float64(progress.elapsed.Nanoseconds()) //bytes por nanosegundo
	progress.remainingTime = int64(float64(progress.totalSize-progress.currentSize) / progress.speed)
	remainingTimeStr := durafmt.Parse(time.Duration(progress.remainingTime)).String()
	speed := progress.speed * 1000000000 * 60 //bytes por minuto
	fmt.Printf("%s/min    %d de %d    %0.2f%%    Estimado: %s\n", humanize.Bytes(uint64(speed)), progress.currentNumber, progress.totalNumber, progress.progressSize, remainingTimeStr)

}
