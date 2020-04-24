package main

import (
	"fmt"
	"os"
	"time"

	"github.com/akamensky/argparse"
)

func main() {

	parser := argparse.NewParser("sinf-mirror", "Mirror one folder to another")
	source := parser.String("s", "source", &argparse.Options{Required: true, Help: "Folder to be mirrored"})
	dest := parser.String("d", "destination", &argparse.Options{Required: true, Help: "Folder to mirror to"})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	synchronizer := new(Synchronizer)
	synchronizer.totalSize = 0
	synchronizer.totalNumber = 0
	synchronizer.currentSize = 0
	synchronizer.source = *source
	synchronizer.dest = *dest
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
