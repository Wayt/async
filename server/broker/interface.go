package broker

import (
	uuid "github.com/satori/go.uuid"
	"github.com/wayt/async/server/job"
)

type Broker interface {
	Consume(JobProcessor) error
	Stop()
	Schedule(*job.Job) error
	List() []*job.Job
	Get(jobID uuid.UUID) (*job.Job, error)
}

type JobProcessor interface {
	Process(*job.Job) (bool, error)
	GetCapabilities() []string
	Stopped() <-chan struct{}
}
