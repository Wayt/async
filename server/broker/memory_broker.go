package broker

import (
	"errors"
	"log"
	"sync"
	"time"

	cache "github.com/patrickmn/go-cache"
	uuid "github.com/satori/go.uuid"
	"github.com/wayt/async/server/job"
)

var (
	ErrJobNotFound = errors.New("job not found")
)

// In memory job broker
type memoryBroker struct {
	sync.Mutex
	stop      chan struct{}
	jobsQueue map[string]chan *job.Job
	jobs      *cache.Cache
}

const jobQueueSize = 200

func NewMemoryBroker() Broker {

	return &memoryBroker{
		stop:      make(chan struct{}),
		jobsQueue: make(map[string]chan *job.Job),
		jobs:      cache.New(5*time.Minute, 10*time.Minute),
	}
}

func (b *memoryBroker) Consume(p JobProcessor) error {

	ch := make(chan *job.Job)
	defer close(ch)

	for _, cap := range p.GetCapabilities() {
		go b.consumeFunc(cap, p.Stopped(), ch)
	}

	for {
		select {
		case <-b.stop:
			return nil
		case <-p.Stopped():
			return nil
		case j := <-ch:
			b.process(p, j)
		}
	}
}

func (b *memoryBroker) process(p JobProcessor, j *job.Job) {
	reschedule, err := p.Process(j)
	if err != nil {
		log.Printf("broker: job process error: %v", err)
	}
	if reschedule {
		if err := b.Schedule(j); err != nil {
			log.Printf("broker: fail to reschedule job [%s][%s]: %v", j.Name, j.ID, err)
		}
	}

}

func (b *memoryBroker) consumeFunc(funcName string, processorStop <-chan struct{}, ch chan *job.Job) {

	q := b.queueForFunc(funcName)

	for {
		select {
		case <-b.stop:
			return
		case <-processorStop:
			log.Printf("broker: stop consumeFunc for %s", funcName)
			return
		case j, ok := <-q:
			if !ok {
				log.Printf("broker: channel closed for func %s", funcName)
				return
			}

			ch <- j
		}
	}
}

func (b *memoryBroker) Stop() {
	close(b.stop)
}

func (b *memoryBroker) queueForFunc(funcName string) chan *job.Job {
	b.Lock()
	defer b.Unlock()

	q, ok := b.jobsQueue[funcName]
	if !ok {
		q = make(chan *job.Job, jobQueueSize)
		b.jobsQueue[funcName] = q
		// log.Printf("broker: created queue %s", funcName)
	}

	return q
}

func (b *memoryBroker) Schedule(j *job.Job) error {

	b.jobs.Add(j.ID.String(), j, cache.DefaultExpiration)

	funcName := j.GetCurrentFunction().Name

	j.ScheduledAt = time.Now()
	q := b.queueForFunc(funcName)
	q <- j

	return nil
}

func (b *memoryBroker) List() []*job.Job {

	jobs := make([]*job.Job, 0, b.jobs.ItemCount())

	for _, item := range b.jobs.Items() {
		jobs = append(jobs, item.Object.(*job.Job))
	}

	return jobs
}

func (b *memoryBroker) Get(jobID uuid.UUID) (*job.Job, error) {

	obj, ok := b.jobs.Get(jobID.String())
	if !ok {
		return nil, ErrJobNotFound
	}

	return obj.(*job.Job), nil
}
