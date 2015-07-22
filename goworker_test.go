package goworker

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"runtime"
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

func TestMain(t *testing.T) {

	quit := make(chan bool, 1)

	// start dispatcher with 10 workers (goroutines) and jobsQueue channel size 20
	d := NewDispatcher(10, 20)

	allJobs := make([]*Job, 20)

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
	go func() {
		for _, j := range allJobs {
			// waiting when job done and then assert
			for !j.Done {
				runtime.Gosched()
			}
			assert.Equal(t, j.SomeJobData, j.SomeJobResult)
		}

		// quit
		quit <- true
	}()

	// start dispatcher
	d.Run(quit)

}
