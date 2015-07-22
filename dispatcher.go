package goworker

import (
	"sync"
)

// Dispatcher starts workers and route jobs for it
type Dispatcher struct {
	// A pool of workers channels that are registered with the goworker
	WorkerPool chan chan GoJob
	Workers    chan *Worker
	maxWorkers int
	jobsQueue  chan GoJob
	quit       chan bool
	wg         *sync.WaitGroup
}

// NewDispatcher construct new Dispatcher
func NewDispatcher(maxWorkers int, jobsQueueSize uint) *Dispatcher {
	jobsQueue := make(chan GoJob, jobsQueueSize)
	pool := make(chan chan GoJob, maxWorkers)
	workers := make(chan *Worker, maxWorkers)
	return &Dispatcher{
		WorkerPool: pool,
		Workers:    workers,
		maxWorkers: maxWorkers,
		jobsQueue:  jobsQueue,
		wg:         &sync.WaitGroup{},
	}
}

// Run dispatcher with quit channel
func (d *Dispatcher) Run(quit chan bool) {
	d.quit = quit

	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		w := NewWorker(i, d.WorkerPool)
		w.Start(d.wg)
		d.Workers <- w
	}

	d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobsQueue:
			// a job request has been received
			go func(job GoJob) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.WorkerPool

				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		case <-d.quit:
			d.stopWorkers()
			return
		}
	}

	d.wg.Wait()
}

func (d *Dispatcher) stopWorkers() {
	for i := 0; i < d.maxWorkers; i++ {
		w := <-d.Workers
		w.Stop()
	}
}

// AddJob adds new job to dispatcher
func (d *Dispatcher) AddJob(job GoJob) {
	d.jobsQueue <- job
}
