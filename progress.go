package main

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hako/durafmt"
)

// Estrutura que armazena informações sobre o progress
type Progress struct {
	totalSize     int64
	totalNumber   int64
	currentSize   int64
	currentNumber int64
	startTime     time.Time
	elapsed       time.Duration
	speed         float64
	remainingTime int64
	progressSize  float64
	newFiles      int64
	updateFiles   int64
	deletedItems  int64
	equalFiles    int64
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
