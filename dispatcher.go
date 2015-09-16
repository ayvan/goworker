//Copyright 2015 Ivan Korostelyov
//
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package goworker

import (
	"sync"
	"time"
)

// Dispatcher starts workers and route jobs for it
type Dispatcher struct {
	// A pool of workers channels that are registered with the goworker
	WorkerPool      chan chan GoJob
	Workers         chan *Worker
	maxWorkers      int
	jobsQueue       chan GoJob
	unperformedJobs []GoJob
	stop            chan bool
	stopped         chan bool
	wg              *sync.WaitGroup
	stopInProgress  bool
}

// NewDispatcher construct new Dispatcher
func NewDispatcher(maxWorkers int, jobsQueueSize uint) *Dispatcher {
	jobsQueue := make(chan GoJob, jobsQueueSize)
	var unperformedJobs []GoJob // when stop copy all jobs from jobsQueue to unperformedJobs
	pool := make(chan chan GoJob, maxWorkers)
	workers := make(chan *Worker, maxWorkers)
	return &Dispatcher{
		WorkerPool:      pool,
		Workers:         workers,
		maxWorkers:      maxWorkers,
		jobsQueue:       jobsQueue,
		unperformedJobs: unperformedJobs,
		stop:            make(chan bool, 1),
		stopped:         make(chan bool, 1),
		wg:              &sync.WaitGroup{},
		stopInProgress:  false,
	}
}

// Run dispatcher
func (d *Dispatcher) Run() {
	// clean
	d.CleanUnperformedJobs()
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		w := NewWorker(i, d.WorkerPool)
		w.Start(d.wg)
		d.Workers <- w
	}

	d.dispatch()
}

func (d *Dispatcher) dispatch() {
	defer func() {
		d.stopWorkers()

		d.wg.Wait()

		// move all jobs from input channel to unperformedJobs
		if len(d.jobsQueue) > 0 {
			for j := range d.jobsQueue {
				d.unperformedJobs = append(d.unperformedJobs, j)
				if len(d.jobsQueue) == 0 {
					break
				}
			}
		}

		d.stopInProgress = false
		d.stopped <- true
	}()

	for {
		select {
		// a job request has been received
		case job := <-d.jobsQueue:
			select {
			// try to obtain a worker job channel that is available.
			// this will block until a worker is idle
			case jobChannel := <-d.WorkerPool:
				// dispatch the job to the worker job channel
				jobChannel <- job
			case <-d.stop:
				// if need to exit save current job to unperformedJobs
				d.unperformedJobs = append(d.unperformedJobs, job)
				return
			}
		case <-d.stop:
			return
		}
	}
}

func (d *Dispatcher) stopWorkers() {
	defer func() {
		// clear WorkerPool
		for _ = range d.WorkerPool {
			if len(d.WorkerPool) == 0 {
				return
			}
		}
	}()

	for w := range d.Workers {
		w.Stop()
		if len(d.Workers) == 0 {
			return
		}
	}
}

// AddJob adds new job to dispatcher
func (d *Dispatcher) AddJob(job GoJob) {
	// add job to dispatcher, it's check is dispatcher Stop() in progress and try again

	// If method d.Stop() called you can't add new jobs before Stop() successfully completed
	// It's required for moving jobs from input channel to unperformedJobs when dispatcher Stop() method called.
	for {
		// loop while stopInProgress
		if d.stopInProgress == false {
			d.jobsQueue <- job
			break
		}
		time.Sleep(time.Microsecond)
	}
}

// Stop dispatcher
func (d *Dispatcher) Stop() {
	d.stopInProgress = true
	d.stop <- true
	<-d.stopped
}

// GetUnperformedJobs method returns a chan of GoJobs that have not been done before Stop() executed
func (d *Dispatcher) GetUnperformedJobs() []GoJob {
	return d.unperformedJobs
}

// CleanUnperformedJobs remove unperformedJobs
func (d *Dispatcher) CleanUnperformedJobs() {
	d.unperformedJobs = make([]GoJob, 0)
}

func (d *Dispatcher) CountJobs() int {
	return len(d.jobsQueue)
}
