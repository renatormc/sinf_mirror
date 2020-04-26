package main

import (
	"sync"
)

// Worker é
type Worker struct {
	synchronizer    *Synchronizer
	results         chan<- ResultData
	jobs            <-chan JobConfig
	acknowledgeDone chan<- bool
	id              int
	wg              *sync.WaitGroup
}

// JobConfig armazena dados sobre o trabalho a ser executado
type JobConfig struct {
	relPath     string
	acknowledge bool // Flag para dizer ao work que ele deve informar que terminou a tarefa
}

func (worker *Worker) init() {
	worker.acknowledgeDone = worker.synchronizer.acknowledgeDone
	worker.jobs = worker.synchronizer.jobs
	worker.results = worker.synchronizer.results
	worker.wg = &worker.synchronizer.wg
}

//Compara arquivo na fonte com destino se for diferente ou não existir copia o novo
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
