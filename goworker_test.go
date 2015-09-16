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
	"fmt"
	"github.com/stretchr/testify/assert"
	"runtime"
	"sync"
	"testing"
)

// Job structure
type Job struct {
	SomeJobData   string
	SomeJobResult string
	Done          bool
}

// check interface implementation
var _ GoJob = (*Job)(nil)

func (j *Job) DoIt() {
	j.SomeJobResult = j.SomeJobData
	j.Done = true
}

func TestGoWorker(t *testing.T) {
	// start dispatcher with 10 workers (goroutines) and jobsQueue channel size 20
	d := NewDispatcher(10, 20)

	allJobs := make([]*Job, 20)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer func() {
			wg.Done()
		}()
		// add jobs to channel
		for i := 0; i < 20; i++ {
			j := &Job{
				SomeJobData: "job number" + fmt.Sprintf("%d", i),
			}
			// put job to slice
			allJobs[i] = j

			// add job to dispatcher
			d.AddJob(j)
		}

		// check is jobs done

		for _, j := range allJobs {
			// waiting when job done and then assert
			for !j.Done {
				runtime.Gosched()
			}
			assert.Equal(t, j.SomeJobData, j.SomeJobResult)

			uj := new(Job)
			*uj = *j
			uj.Done = false
			// for testing unperformed jobs methods
			d.unperformedJobs = append(d.unperformedJobs, uj)
		}

		// count jobs
		assert.Equal(t, 0, d.CountJobs())

		// stop dispatcher
		d.Stop()

		// copy jobs
		uJ := d.GetUnperformedJobs()

		// assert len of jobs
		assert.Equal(t, 20, len(d.GetUnperformedJobs()))

		// clean jobs
		d.CleanUnperformedJobs()

		// assert len of jobs in dispatcher and in slice
		assert.Equal(t, 0, len(d.GetUnperformedJobs()))
		assert.Equal(t, 20, len(uJ))

		// add unperformed jobs again
		for i, job := range uJ {
			allJobs[i] = job.(*Job)

			// simulate highload to stop dispatcher when jobs queue not empty
			for i := 0; i < 1000; i++ {
				go d.AddJob(job)
			}
		}

		go d.Run()
		d.Stop()
	}()

	// start dispatcher
	d.Run()

	wg.Wait()
}
