package main

import (
	"fmt"
	"time"
)

func main() {
	synchronizer := new(Synchronizer)
	synchronizer.totalSize = 0
	synchronizer.totalNumber = 0
	synchronizer.currentSize = 0
	synchronizer.source = "D:\\teste_report\\C3"
	synchronizer.dest = "D:\\teste_report\\c1_copia_deletar"
	synchronizer.verbose = true
	synchronizer.startTime = time.Now()
	synchronizer.appConfig = GetConfig()

	fmt.Println("Contando arquivos...")
	synchronizer.countFiles()
	fmt.Printf("Quantidade total: %d, Tamanho total: %d \n", synchronizer.totalNumber, synchronizer.totalSize)
	fmt.Println("Sincronizando...")
	synchronizer.update()

	fmt.Printf("Tempo gasto: %s\n", synchronizer.elapsed)
	fmt.Println(synchronizer.totalSize, synchronizer.totalNumber)
}
