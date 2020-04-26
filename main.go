package main

import (
	"fmt"
	"os"
	"time"

	"github.com/akamensky/argparse"
	"github.com/dustin/go-humanize"
	"github.com/hako/durafmt"
)

type Config struct {
	source     string        //pasta fonte
	dest       string        // pasta destino
	verbose    bool          // Imprimir mensagens extras
	NWorkers   int           // Número de workers
	threshold  int64         // Tamanho em megabytes a partir do qual será copiado sem concorrência
	bufferSize int64         // Tamanho do buffer utilizado para copiar arquivos grandes
	purge      bool          // Deletar o que existir no destino e não na fonte
	retries    int           // Número de tentativas de copiar arquivo caso ocorra erros
	wait       time.Duration // Tempo para aguardar antes de tentar de novo em segundos
}

func main() {

	parser := argparse.NewParser("sinf-mirror", "Mirror one folder to another")
	source := parser.String("s", "source", &argparse.Options{Required: true, Help: "Folder to be mirrored"})
	dest := parser.String("d", "destination", &argparse.Options{Required: true, Help: "Folder to mirror to"})
	nWorkers := parser.Int("w", "workers", &argparse.Options{Default: 8, Help: "Number of workers"})
	threshold := parser.Int("t", "threshold", &argparse.Options{Default: 8, Help: "Size above which there will be no concurrency"})
	bufferSize := parser.Int("b", "buffer", &argparse.Options{Default: 1, Help: "Buffer size in megabytes"})
	verbose := parser.Flag("v", "verbose", &argparse.Options{Default: false, Help: "Verbose"})
	purge := parser.Flag("p", "purge", &argparse.Options{Default: false, Help: "Purge"})
	retries := parser.Int("r", "retries", &argparse.Options{Default: 10, Help: "Specifies the number of retries on failed copies"})
	waitTime := parser.Int("t", "wait", &argparse.Options{Default: 30, Help: "Specifies the wait time between retries, in seconds."})
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
	config.bufferSize = int64(*bufferSize) * 1048576
	config.purge = *purge
	config.retries = *retries
	config.threshold = *threshold
	config.wait = time.Duration((*waitTime) * 1000000000)
	fmt.Println("Contando arquivos...")
	var progress Progress
	progress.totalSize = 0
	var counter Counter
	counter.config = &config
	counter.countFiles()
	progress.totalNumber = counter.totalNumber
	progress.totalSize = counter.totalSize
	fmt.Printf("Número de arquivos encontrados: %d\n", progress.totalNumber)
	fmt.Printf("Total bytes: %s\n", humanize.Bytes(uint64(progress.totalSize)))
	progress.newFiles = 0
	progress.updateFiles = 0
	progress.deletedItems = 0

	jobs := make(chan WorkerConfig)
	results := make(chan ResultData)
	finished := make(chan bool)
	go progressWork(&config, &progress, results, finished)
	fmt.Println("Sincronizando...")
	progress.startTime = time.Now()
	go update(&config, jobs, results)

	<-finished

	fmt.Println("\nProcesso finalizado")
	fmt.Printf("\nArquivos analisados: %d\n", progress.totalNumber)
	fmt.Printf("Arquivos novos: %d\n", progress.newFiles)
	fmt.Printf("Arquivos atualizados: %d\n", progress.updateFiles)
	fmt.Printf("Arquivos iguais: %d\n", progress.equalFiles)
	fmt.Printf("Itens deletados: %d\n", progress.deletedItems)
	fmt.Printf("Número de workers: %d\n", config.NWorkers)
	elapsed := durafmt.Parse(time.Duration(progress.elapsed)).String()
	fmt.Printf("Tempo gasto: %s\n", elapsed)
	speed := progress.speed * 1000000000 * 60 //bytes por minuto
	fmt.Printf("Velocidade média: %s/min\n", humanize.Bytes(uint64(speed)))
}
