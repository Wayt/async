package broker

import (
	"log"
	"sync"
	"time"

	"github.com/wayt/async/server/job"
)

// In memory job broker
type memoryBroker struct {
	lock     *sync.Mutex
	stop     chan struct{}
	jobQueue chan *job.Job
}

const jobQueueSize = 200

func NewMemoryBroker() Broker {

	return &memoryBroker{
		stop:     make(chan struct{}),
		jobQueue: make(chan *job.Job, jobQueueSize),
	}
}

func (b *memoryBroker) Consume(jobProcessor JobProcessor) error {

	for {
		select {
		case <-b.stop:
			return nil
		case j := <-b.jobQueue:
			reschedule, err := jobProcessor.Process(j)
			if err != nil {
				log.Printf("broker: job process error: %v", err)
			}
			if reschedule {
				if err := b.Schedule(j); err != nil {
					log.Printf("broker: fail to reschedule job [%s][%s]: %v", j.Name, j.ID, err)
				}
			}
		}
	}
}

func (b *memoryBroker) Stop() {
	close(b.stop)
}

func (b *memoryBroker) Schedule(j *job.Job) error {
	j.ScheduledAt = time.Now()
	b.jobQueue <- j
	return nil
}
