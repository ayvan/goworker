package goworker

// GoJob is interface for jobs
type GoJob interface {
	DoIt()
}