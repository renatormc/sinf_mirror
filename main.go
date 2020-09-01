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
	inputFile := parser.String("f", "input-file", &argparse.Options{Default: "null", Help: "Input file with sources and destination"})
	caseName := parser.String("c", "casename", &argparse.Options{Default: "null", Help: "Case name"})
	nWorkers := parser.Int("w", "workers", &argparse.Options{Default: 10, Help: "Number of workers"})
	maxDepth := parser.Int("m", "max-depth", &argparse.Options{Default: 5, Help: "Max depth to scan for case folders"})
	threshold := parser.Int("t", "threshold", &argparse.Options{Default: 8, Help: "Size in megabytes above which there will be no concurrency"})
	thresholdChunk := parser.Int("k", "threshold-chunk", &argparse.Options{Default: 8388600, Help: "Size in megabytes above which file will be copied in chunks"})
	bufferSize := parser.Int("b", "buffer", &argparse.Options{Default: 1, Help: "Buffer sisfgze in megabytes"})
	verbose := parser.Flag("v", "verbose", &argparse.Options{Default: false, Help: "Verbose"})
	purge := parser.Flag("p", "purge", &argparse.Options{Default: false, Help: "Purge"})
	retries := parser.Int("r", "retries", &argparse.Options{Default: 10, Help: "Specifies the number of retries on failed copies"})
	waitTime := parser.Int("i", "wait", &argparse.Options{Default: 1, Help: "Specifies the wait time between retries, in seconds."})
	logging := parser.String("l", "log", &argparse.Options{Default: "null", Help: "Logs into specified file. A separate disk or thumbdrive is recommend. It is not recommended to use the same disk as source or desitnation due to performance drops"})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	var synchronizer Synchronizer
	synchronizer.maxDepth = *maxDepth
	synchronizer.caseName = *caseName
	synchronizer.init()
	if *inputFile != "null" {
		synchronizer.autoFind = true
		synchronizer.sources, synchronizer.dest = getInputsFromFile(*inputFile)

	} else if synchronizer.caseName != "null" {
		fmt.Printf("Sincronizar pastas do caso %s\n", synchronizer.caseName)
		synchronizer.autoFind = true
		synchronizer.scanDrives()
		if len(synchronizer.sources) == 0 {
			fmt.Printf("Não foi encontrada nenhuma pasta temp relacionada ao caso \"%s\"\n", synchronizer.caseName)
			os.Exit(1)
		}
		if synchronizer.dest == "" {
			fmt.Printf("Não foi encontrada nenhuma pasta final relacionada ao caso \"%s\"\n", synchronizer.caseName)
			os.Exit(1)
		}

	} else {
		synchronizer.sources = *sources
		synchronizer.dest = *dest
		synchronizer.autoFind = false
	}
	synchronizer.foldersOK()
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

	//initializing logger
	if *logging == "null" {
		fmt.Printf("Logging is not enabled\n")
		synchronizer.logging = false
	} else {
		fmt.Printf("Logging into folder %s\n", *logging)
		synchronizer.logging = true
		var l Logger                             // creates  logger variable
		synchronizer.logger = l                  //sets synchronizer logger variable
		synchronizer.logger.logOutput = *logging //sets logging folder
		synchronizer.logger.init()
		if synchronizer.logger.logging {
			go synchronizer.logger.run() //launches logger

		} else {
			*logging = "null"            //disables MAIN logging flag
			synchronizer.logging = false //disables syncronyzer logging flag
		}

	}

	go progress.run()
	go synchronizer.run()

	<-progress.finished

	//waiting logging operations to be finished
	if *logging != "null" {
		<-synchronizer.logger.finishedLogging
		fmt.Println("Logging operations finished")
	}

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
