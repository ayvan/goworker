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

func TestGoWorker(t *testing.T) {
	// start dispatcher with 10 workers (goroutines) and jobsQueue channel size 20
	d := NewDispatcher(10, 20)

	allJobs := make([]*Job, 20)

	go func() {
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

			// for testing unperformed jobs methods
			d.unperformedJobs = append(d.unperformedJobs, j)
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
		for _, job := range uJ {
			d.AddJob(job)
		}

		// count jobs
		assert.Equal(t, 20, d.CountJobs())
	}()

	// start dispatcher
	d.Run()

}
