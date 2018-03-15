package broker

import "github.com/wayt/async/server/job"

type Broker interface {
	Consume(JobProcessor) error
	Stop()
	Schedule(*job.Job) error
}

type JobProcessor interface {
	Process(*job.Job) (bool, error)
}
