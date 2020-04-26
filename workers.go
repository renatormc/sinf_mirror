package main

import (
	"sync"
)

// Worker mantém as configurações do worker
type Worker struct {
	synchronizer    *Synchronizer     // Ponteiro para o sincronizador
	results         chan<- ResultData // Canal utilizado para passar o resultado da tarefa par ao gerenciador de progresso
	jobs            <-chan JobConfig  // Canal por onde o worker receberá as tarefas
	acknowledgeDone chan<- bool       // Canal utilizado para informar o sincronizador que foi terminada a tarefa de copia de um arquivo grande
	id              int               // Número inteiro que indentifica o woker cada um tem o seu
	wg              *sync.WaitGroup   // Ponteiro para o objeto que é reponsável por garantir que todos os workers terminaram suas tarefas
}

// JobConfig armazena dados sobre o trabalho a ser executado
type JobConfig struct {
	relPath     string // Caminho relativo do arquivo
	acknowledge bool   // Flag para dizer ao work que ele deve informar que terminou a tarefa
}

func (worker *Worker) init() {
	worker.acknowledgeDone = worker.synchronizer.acknowledgeDone
	worker.jobs = worker.synchronizer.jobs
	worker.results = worker.synchronizer.results
	worker.wg = &worker.synchronizer.wg
}

// Função de execução do worker vai pegando as tarefas que forem aparecendo no canal jobs e executando.
// Caso a flag acknowledge estiver setada ele informa através do canal acknowledgeDone o sincronizador que terminou a tarefa.
// O Sincronizador que escolhe quando o woker deve informar que terminou e quando não precisa. Normalmente será preciso somente quando o arquivo for grande.
// Nesse caso o sincronizador não irá atribuir tarefa aos outros workers enquanto não for informado que a cópia do arquivo grande foi finalizada
func (worker *Worker) run() {

	defer func() {
		worker.wg.Done()
	}()
	for j := range worker.jobs {
		if j.relPath != "wfinish" {
			worker.synchronizer.copyOrReplaceFile(j.relPath)
			if j.acknowledge {
				worker.acknowledgeDone <- true
			}
		} else {
			return
		}
	}

}
