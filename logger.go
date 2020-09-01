package main

import (
	"fmt"
	"os"
	"time"
)

//Logger is a struct with associated methods that will receive, buffer, flush and control logging operations
type Logger struct {
	logBufferSize   int
	logResults      []LogInfo    //logger buffer
	receiver        chan LogInfo //chanel to receive logging messages
	finishedLogging chan bool    //channel to send informatition about
	logOutput       string
	logging         bool
	logFile         *os.File
}

type LogInfo struct {
	logPath   string
	logAction string
}

func (logger *Logger) outputOK() bool {

	_, err := os.Stat(logger.logOutput)
	if os.IsNotExist(err) {
		fmt.Printf("O diretório \"%s\" não existe.", logger.logOutput)
		return false
	} else {
		fmt.Println("Logging into directory " + logger.logOutput)
	}

	return true

}

func (logger *Logger) init() {

	logger.logBufferSize = 104857600 //dev define buffer size 104.857.600  100Kib
	logger.logResults = make([]LogInfo, 0, logger.logBufferSize)
	fmt.Printf("Logging enabled with bufer size = %d \n", logger.logBufferSize)
	logger.finishedLogging = make(chan bool)
	logger.receiver = make(chan LogInfo)

	//for testing purposes only
	//logger.logOutput = "C:/Users/Will/Desktop/logOutput"
	if logger.outputOK() {
		var err error
		logger.logFile, err = os.Create(logger.logOutput + "/Mirror_log.log")
		if err != nil {
			logger.logging = false
			fmt.Println("Error while opening file and log was turned off ") // ]
			//FUTURAMENTE mudar isso aqui p/ perguntar o q o usuário quer fazer. Abortar ou continuar sem log.
		} else {
			fmt.Println("Logging is active - file is openned")
			logger.logging = true
		}

		//defer logger.logFile.Close()

	} else {
		fmt.Println("Logging output is not ok")
		logger.logging = false
	}
}

func (logger *Logger) flushLog() {
	if logger.logBufferSize < 1 || logger.logResults == nil {
		return
	}
	//parei aqui
	for _, item := range logger.logResults {
		//fmt.Println("Flushing item " + item.logAction + item.logPath)

		//code to log stuff comes here
		if logger.logging {

			logger.logFile.WriteString(item.logAction + " " + item.logPath + "\n")
		}

		//end code to log stuff

	}
	if logger.logging {

		logger.logFile.Sync()
	}

	logger.logResults = make([]LogInfo, 0, logger.logBufferSize)
	//r.Printf("Newly create buffer len = %d\n", len(logger.logResults))

}

// Função de execução do worker vai pegando as tarefas que forem aparecendo no canal jobs e executando.
// Caso a flag acknowledge estiver setada ele informa através do canal acknowledgeDone o sincronizador que terminou a tarefa.
// O Sincronizador que escolhe quando o woker deve informar que terminou e quando não precisa. Normalmente será preciso somente quando o arquivo for grande.
// Nesse caso o sincronizador não irá atribuir tarefa aos outros workers enquanto não for informado que a cópia do arquivo grande foi finalizada
func (logger *Logger) run() {
	fmt.Printf("Initializing logger\n")

	for item := range logger.receiver {

		if item.logAction == "!@#finish" {
			fmt.Println("Secret message!")
			if len(logger.logResults) > 0 {
				logger.flushLog()
			}
			break
		}

		//fmt.Printf("Received log message %s\n", item) //debbug purposes only
		//fmt.Printf("Buffer capacity = %d Buffer length = %d \n", logger.logBufferSize, len(logger.logResults))

		//add element to resultBuffer
		logger.logResults = append(logger.logResults, item)

		if len(logger.logResults) >= logger.logBufferSize {
			logger.flushLog()
		}

	}
	time.Sleep(time.Second)
	logger.finishedLogging <- true
	if logger.logging {
		logger.logFile.Close()
	}

	defer func() {
		//worker.wg.Done()
	}()
}
