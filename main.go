package main

import (
	"fmt"
	"os"
	"time"

	"github.com/akamensky/argparse"
	"github.com/dustin/go-humanize"
)

func main() {

	parser := argparse.NewParser("sinf-mirror", "Mirror one folder to another")
	sources := parser.StringList("s", "source", &argparse.Options{Help: "Folder to be mirrored"})
	dest := parser.String("d", "destination", &argparse.Options{Help: "Folder to mirror to"})
	caseName := parser.String("c", "casename", &argparse.Options{Default: "!!", Help: "Case name"})
	maxDeep := parser.Int("m", "max-deep", &argparse.Options{Default: 3, Help: "How deep the program search for case folders"})
	nWorkers := parser.Int("w", "workers", &argparse.Options{Default: 10, Help: "Number of workers"})
	threshold := parser.Int("t", "threshold", &argparse.Options{Default: 8, Help: "Size in megabytes above which there will be no concurrency"})
	thresholdChunk := parser.Int("k", "threshold-chunk", &argparse.Options{Default: 8388600, Help: "Size in megabytes above which file will be copied in chunks"})
	bufferSize := parser.Int("b", "buffer", &argparse.Options{Default: 1, Help: "Buffer size in megabytes"})
	verbose := parser.Flag("v", "verbose", &argparse.Options{Default: false, Help: "Verbose"})
	purge := parser.Flag("n", "purge", &argparse.Options{Default: false, Help: "Purge"})
	retries := parser.Int("r", "retries", &argparse.Options{Default: 10, Help: "Specifies the number of retries on failed copies"})
	waitTime := parser.Int("i", "wait", &argparse.Options{Default: 1, Help: "Specifies the wait time between retries, in seconds."})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	var synchronizer Synchronizer
	synchronizer.init()
	if *caseName == "!!" {
		synchronizer.sources = *sources
		synchronizer.dest = *dest
		synchronizer.autoFind = false
	} else {
		synchronizer.autoFind = true
		folderAnalyzer := FolderAnalyzer{MaxDeep: *maxDeep}
		folderAnalyzer.init()
		fmt.Println("Procurando pastas do caso...")
		folderAnalyzer.findFolders(*caseName)
		if len(folderAnalyzer.Sources) == 0 {
			fmt.Println("Não foi encontrada nenhuma pasta com dados do caso a serem copiados.")
			os.Exit(1)
		}
		if folderAnalyzer.Destination == "" {
			fmt.Println("A pasta de destino não foi encontrada.")
			os.Exit(1)
		}
		synchronizer.sources = folderAnalyzer.Sources
		synchronizer.dest = folderAnalyzer.Destination
		// fmt.Println(synchronizer.sources, synchronizer.dest)
		// os.Exit(0)
	}
	synchronizer.verbose = *verbose
	synchronizer.NWorkers = *nWorkers
	synchronizer.bufferSize = int64(*bufferSize) * 1048576
	synchronizer.thresholdChunk = int64(*thresholdChunk) * 1048576
	synchronizer.purge = *purge
	synchronizer.retries = *retries
	synchronizer.threshold = int64(*threshold) * 1048576
	synchronizer.wait = time.Duration((*waitTime) * 1000000000)
	fmt.Println("Contando arquivos...")
	progress := Progress{synchronizer: &synchronizer}
	progress.init()
	progress.countFiles()
	fmt.Printf("N arquivos encontrados: %d\n", progress.totalNumber)
	fmt.Printf("Total bytes: %s\n", humanize.Bytes(uint64(progress.totalSize)))

	fmt.Println("Sincronizando...")
	go progress.run()
	go synchronizer.run()

	<-progress.finished

	fmt.Printf("\nTamanho total:           %s\n", humanize.Bytes(uint64(progress.totalSize)))
	fmt.Printf("Arquivos analisados:     %d\n", progress.totalNumber)
	fmt.Printf("Arquivos novos:          %d\n", progress.newFiles)
	fmt.Printf("Arquivos atualizados:    %d\n", progress.updatedFiles)
	fmt.Printf("Arquivos iguais:         %d\n", progress.equalFiles)
	fmt.Printf("Itens deletados:         %d\n", progress.deletedItems)
	fmt.Printf("N workers:               %d\n", synchronizer.NWorkers)
	elapsed := fmtDuration(time.Duration(progress.elapsed))
	fmt.Printf("Tempo gasto:             %s\n", elapsed)
	speed := progress.avgSpeed * 1000000000 * 60 //bytes por minuto
	fmt.Printf("Velocidade média:        %s/min\n", humanize.Bytes(uint64(speed)))
}
