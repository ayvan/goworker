# goworker

[![Build Status](https://travis-ci.org/Ayvan/goworker.svg?branch=master)](https://travis-ci.org/Ayvan/goworker)

Based on http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/

### Example:

```
package main

import (
	"fmt"
	"goworker"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Job struct {
	SomeJobData string
}

// check interface implementation
var _ goworker.GoJob = (*Job)(nil)

// DoIt - Job struct must implements GoJob interface, need this function
func (j *Job) DoIt() {
	log.Println("Worker do this job: ", j.SomeJobData)
}

func main() {

	//init system signal channel, need for Ctrl+C and, kill and other system signal handling

	sChan := make(chan os.Signal, 1)
	signal.Notify(sChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	// start dispatcher with 10 workers (goroutines) and jobsQueue channel size 20
	d := goworker.NewDispatcher(10, 20)

	go func() {
		// catch quit signal
		s := <-sChan
		log.Println("os.Signal", s, "received, finishing application...")
		// stop dispatcher
		d.Stop()
		return
	}()

	go func() {
		for i := 1; i < 100; i++ {
			j := &Job{
				SomeJobData: "job number" + fmt.Sprintf("%d", i),
			}

			// add job to dispatcher
			d.AddJob(j)
		}
	}()

	// start dispatcher
	d.Run()
}
```