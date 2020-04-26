package main

import (
	"fmt"
	"os"
	"time"

	"github.com/akamensky/argparse"
	"github.com/dustin/go-humanize"
	"github.com/hako/durafmt"
)

func main() {

	parser := argparse.NewParser("sinf-mirror", "Mirror one folder to another")
	source := parser.String("s", "source", &argparse.Options{Required: true, Help: "Folder to be mirrored"})
	dest := parser.String("d", "destination", &argparse.Options{Required: true, Help: "Folder to mirror to"})
	nWorkers := parser.Int("w", "workers", &argparse.Options{Default: 8, Help: "Number of workers"})
	threshold := parser.Int("t", "threshold", &argparse.Options{Default: 8, Help: "Size in megabytes above which there will be no concurrency"})
	bufferSize := parser.Int("b", "buffer", &argparse.Options{Default: 1, Help: "Buffer size in megabytes"})
	verbose := parser.Flag("v", "verbose", &argparse.Options{Default: false, Help: "Verbose"})
	purge := parser.Flag("p", "purge", &argparse.Options{Default: false, Help: "Purge"})
	retries := parser.Int("r", "retries", &argparse.Options{Default: 10, Help: "Specifies the number of retries on failed copies"})
	waitTime := parser.Int("i", "wait", &argparse.Options{Default: 1, Help: "Specifies the wait time between retries, in seconds."})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	var synchronizer Synchronizer
	synchronizer.init()
	synchronizer.source = *source
	synchronizer.dest = *dest
	synchronizer.verbose = *verbose
	synchronizer.NWorkers = *nWorkers
	synchronizer.bufferSize = int64(*bufferSize) * 1048576
	synchronizer.purge = *purge
	synchronizer.retries = *retries
	synchronizer.threshold = int64(*threshold) * 1048576
	synchronizer.wait = time.Duration((*waitTime) * 1000000000)
	fmt.Println("Contando arquivos...")
	progress := Progress{synchronizer: &synchronizer}
	progress.init()
	progress.countFiles()
	fmt.Printf("Número de arquivos encontrados: %d\n", progress.totalNumber)
	fmt.Printf("Total bytes: %s\n", humanize.Bytes(uint64(progress.totalSize)))

	fmt.Println("Sincronizando...")
	go progress.run()
	go synchronizer.run()

	<-progress.finished

	fmt.Println("\nProcesso finalizado")
	fmt.Printf("\nArquivos analisados: %d\n", progress.totalNumber)
	fmt.Printf("Arquivos novos: %d\n", progress.newFiles)
	fmt.Printf("Arquivos atualizados: %d\n", progress.updateFiles)
	fmt.Printf("Arquivos iguais: %d\n", progress.equalFiles)
	fmt.Printf("Itens deletados: %d\n", progress.deletedItems)
	fmt.Printf("Número de workers: %d\n", synchronizer.NWorkers)
	elapsed := durafmt.Parse(time.Duration(progress.elapsed)).String()
	fmt.Printf("Tempo gasto: %s\n", elapsed)
	speed := progress.speed * 1000000000 * 60 //bytes por minuto
	fmt.Printf("Velocidade média: %s/min\n", humanize.Bytes(uint64(speed)))
}
