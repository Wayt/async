package async

import (
	"errors"
	"fmt"
	"log"
	"time"

	cache "github.com/pmylund/go-cache"
	"github.com/satori/go.uuid"
)

const (
	StatePending = "pending"
	StateDoing   = "doing"
	StateDone    = "done"
	StateFailed  = "failed"
)

var (
	ErrNotFound   = errors.New("not found")
	errReschedule = errors.New("reschedule")
	errAbort      = errors.New("abort")
)

type Job struct {
	ID              uuid.UUID              `json:"job_id"`
	Name            string                 `json:"name"`
	Functions       []*Function            `json:"functions"`
	CurrentFunction int                    `json:"current_step"`
	State           string                 `json:"state"`
	Data            map[string]interface{} `json:"data"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// JobRepository

type JobRepository interface {
	Get(id uuid.UUID) (*Job, error)
	Create(*Job) (*Job, error)
	Update(*Job) (*Job, error)
	Delete(id interface{}) error

	GetPending() (*Job, bool)
	Schedule(*Job) error
}

// In memory job repository
type memoryJobRepository struct {
	db       *cache.Cache
	jobQueue chan *Job
}

const jobQueueSize = 200

func NewJobRepository() JobRepository {

	return &memoryJobRepository{
		db:       cache.New(cache.NoExpiration, time.Minute),
		jobQueue: make(chan *Job, jobQueueSize),
	}
}

func (r *memoryJobRepository) Get(id uuid.UUID) (*Job, error) {

	i, ok := r.db.Get(id.String())
	if !ok {

		return nil, ErrNotFound
	}

	return i.(*Job), nil
}

func (r *memoryJobRepository) Create(j *Job) (*Job, error) {

	var err error
	if j.ID, err = uuid.NewV4(); err != nil {
		return nil, err
	}

	j.CreatedAt = time.Now()
	r.db.Set(j.ID.String(), j, cache.NoExpiration)

	return j, r.Schedule(j)
}

func (r *memoryJobRepository) Schedule(j *Job) error {
	r.jobQueue <- j
	return nil
}

func (r *memoryJobRepository) Update(j *Job) (*Job, error) {

	j.UpdatedAt = time.Now()

	r.db.Set(j.ID.String(), j, cache.NoExpiration)

	return j, nil
}

func (r *memoryJobRepository) Delete(id interface{}) error {

	r.db.Delete(fmt.Sprintf("%v", id))

	return nil
}

func (r *memoryJobRepository) GetPending() (*Job, bool) {

	j, ok := <-r.jobQueue
	return j, ok
}

type JobManager struct {
	repository       JobRepository
	functionExecutor FunctionExecutor
}

func NewJobManager(repo JobRepository, fexec FunctionExecutor) *JobManager {

	return &JobManager{
		repository:       repo,
		functionExecutor: fexec,
	}
}

func (m *JobManager) GetByID(id uuid.UUID) (*Job, error) {
	return m.repository.Get(id)
}

func (m *JobManager) Create(name string, functions []*Function, data map[string]interface{}) (*Job, error) {

	if len(functions) == 0 {
		return nil, fmt.Errorf("cannot create a job with empty functions")
	}

	j := &Job{
		Name:            name,
		Functions:       functions,
		State:           StatePending,
		Data:            data,
		CurrentFunction: 0,
	}

	return m.repository.Create(j)
}

func (m *JobManager) Execute(j *Job) error {
	var err error

	j.State = StateDoing
	if j, err = m.repository.Update(j); err != nil {
		return err
	}

	for i := range j.Functions {
		if i != j.CurrentFunction {
			log.Printf("Skip job step %s(%d), current step is %s(%d)", j.Functions[i].Name, i, j.Functions[j.CurrentFunction].Name, j.CurrentFunction)
			continue
		}

		// Execute function
		if err = m.executeCurrentFunction(j); err != nil {
			break
		}

		j.CurrentFunction += 1
		if j, err = m.repository.Update(j); err != nil {
			break
		}
	}

	switch err {
	case errReschedule:
		return m.reschedule(j)
	case errAbort:
		return m.abort(j)
	case nil:
		return m.done(j)
	default:
		return err
	}
}

func (m *JobManager) reschedule(j *Job) error {
	var err error

	log.Printf("reschedule %s - %s", j.ID, j.Name)

	j.State = StatePending
	if j, err = m.repository.Update(j); err != nil {
		return err
	}

	return m.repository.Schedule(j)
}

func (m *JobManager) abort(j *Job) error {
	var err error

	log.Printf("abort %s - %s", j.ID, j.Name)

	j.State = StateFailed
	if j, err = m.repository.Update(j); err != nil {
		return err
	}

	return nil
}
func (m *JobManager) done(j *Job) error {
	var err error

	log.Printf("done %s - %s", j.ID, j.Name)

	j.State = StateDone
	if j, err = m.repository.Update(j); err != nil {
		return err
	}

	return nil
}

func (m *JobManager) executeCurrentFunction(j *Job) error {
	var err error

	f := j.Functions[j.CurrentFunction]
	data := j.Data

	f.RetryCount += 1
	if j, err = m.repository.Update(j); err != nil {
		return err
	}

	if err = m.functionExecutor.Execute(*f, data); err != nil {
		if f.CanReschedule() {
			log.Printf("function %s failed, rescheduling", f.Name)
			return errReschedule
		} else {
			log.Printf("function %s failed, cannot reschedule", f.Name)
			return errAbort
		}
	}

	return nil
}

func (m *JobManager) BackgroundProcess() {

	for {
		j, ok := m.repository.GetPending()
		if !ok {
			log.Printf("Stop processing job...")
			return
		}

		log.Printf("Processing job %s", j.ID)
		if err := m.Execute(j); err != nil {
			log.Printf("BackgroundProcess error: %v", err)
		}
	}
}
