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
)

// Worker represents the worker that executes the job
type Worker struct {
	WorkerPool chan chan GoJob
	JobChannel chan GoJob
	quit       chan bool
	num        int
}

// NewWorker constructs new Worker
func NewWorker(num int, workerPool chan chan GoJob) *Worker {
	return &Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan GoJob),
		quit:       make(chan bool),
		num:        num,
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
