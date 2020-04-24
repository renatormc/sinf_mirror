package main

import (
	"fmt"
	"os"
	"time"

	"github.com/akamensky/argparse"
	"github.com/hako/durafmt"
)

type Config struct {
	source   string
	dest     string
	verbose  bool
	NWorkers int
}

func main() {

	parser := argparse.NewParser("sinf-mirror", "Mirror one folder to another")
	source := parser.String("s", "source", &argparse.Options{Required: true, Help: "Folder to be mirrored"})
	dest := parser.String("d", "destination", &argparse.Options{Required: true, Help: "Folder to mirror to"})
	nWorkers := parser.Int("w", "workers", &argparse.Options{Default: 2, Help: "Number of workers"})
	verbose := parser.Flag("v", "verbose", &argparse.Options{Default: false, Help: "Verbose"})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	var config Config
	config.source = *source
	config.dest = *dest
	config.verbose = *verbose
	config.NWorkers = *nWorkers

	fmt.Println("Contando arquivos...")
	var progress Progress
	progress.totalSize = 0
	var counter Counter
	counter.config = &config
	counter.countFiles()
	progress.totalNumber = counter.totalNumber
	progress.totalSize = counter.totalSize
	progress.newFiles = 0
	progress.updateFiles = 0
	progress.deletedItems = 0
	progress.startTime = time.Now()

	jobs := make(chan WorkerConfig)
	results := make(chan ResultData)
	finished := make(chan bool)
	go progressWork(&config, &progress, results, finished)
	fmt.Println("Sincronizando...")
	go update(&config, jobs, results)

	<-finished

	fmt.Println("\nProcesso finalizado")
	fmt.Printf("\nArquivos analisados: %d\n", progress.totalNumber)
	fmt.Printf("Arquivos novos: %d\n", progress.newFiles)
	fmt.Printf("Arquivos atualizados: %d\n", progress.updateFiles)
	fmt.Printf("Itens deletados: %d\n", progress.deletedItems)
	fmt.Printf("Número de workers: %d\n", config.NWorkers)
	elapsed := durafmt.Parse(time.Duration(progress.elapsed)).String()
	fmt.Printf("Tempo gasto: %s\n", elapsed)
}
