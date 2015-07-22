package goworker

import (
	"sync"
)

// Worker represents the worker that executes the job
type Worker struct {
	WorkerPool chan chan GoJob
	JobChannel chan GoJob
	quit       chan bool
	num        int
}

// NewWorker constructs new Worker
func NewWorker(num int, workerPool chan chan GoJob) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan GoJob),
		quit:       make(chan bool),
		num: num,
	}
}

// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w *Worker) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// register the current worker into the worker queue.
			w.WorkerPool <- w.JobChannel
			select {
				case job := <-w.JobChannel:
			// we have received a work request.
				job.DoIt()

			case <-w.quit:
				// we have received a signal to stop
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}
