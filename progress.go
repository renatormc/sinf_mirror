package main

import (
	"fmt"
	"time"

	"github.com/hako/durafmt"
)

// Estrutura que armazena informações sobre o progress
type Progress struct {
	totalSize      int64
	totalNumber    int64
	currentSize    int64
	currentNumber  int64
	startTime      time.Time
	elapsed        time.Duration
	speed          float64
	remainingTime  int64
	progressSize   float64
	progressNumber float64
	newFiles       int64
	updateFiles    int64
	deletedItems   int64
}

// Calcula o progresso
func (progress *Progress) calculateProgress() {
	progress.progressNumber = 100 * float64(progress.currentNumber) / float64(progress.totalNumber)
	progress.progressSize = 100 * float64(progress.currentSize) / float64(progress.totalSize)
	progress.elapsed = time.Since(progress.startTime)
	progress.speed = float64(progress.currentSize) / float64(progress.elapsed.Nanoseconds()) //bytes por nanosegundo
	progress.remainingTime = int64(float64(progress.totalSize-progress.currentSize) / progress.speed)
	remainingTimeStr := durafmt.Parse(time.Duration(progress.remainingTime)).String()
	fmt.Printf("N Items analizados: %0.2f%%		 Bytes: %0.2f%%		 Estimado: %s\n", progress.progressNumber, progress.progressSize, remainingTimeStr)
}
