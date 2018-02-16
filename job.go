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

type JobManager interface {
	GetByID(uuid.UUID) (*Job, error)
	Create(string, []*Function, map[string]interface{}) (*Job, error)

	SetState(*Job, string) error
	Reschedule(*Job) error
	IncrCurrentFunctionRetryCount(*Job) error
	IncrCurrentFunction(*Job) error
}

type jobManager struct {
	repository JobRepository
}

func NewJobManager(repo JobRepository) *jobManager {

	return &jobManager{
		repository: repo,
	}
}

func (m *jobManager) GetByID(id uuid.UUID) (*Job, error) {
	return m.repository.Get(id)
}

func (m *jobManager) Create(name string, functions []*Function, data map[string]interface{}) (*Job, error) {

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

func (m *jobManager) SetState(j *Job, state string) error {
	var err error

	j.State = state
	if j, err = m.repository.Update(j); err != nil {
		return err
	}

	return nil
}

func (m *jobManager) Reschedule(j *Job) error {

	log.Printf("reschedule %s - %s", j.ID, j.Name)

	if err := m.SetState(j, StatePending); err != nil {
		return err
	}

	return m.repository.Schedule(j)
}

func (m *jobManager) IncrCurrentFunctionRetryCount(j *Job) error {
	var err error

	j.Functions[j.CurrentFunction].RetryCount += 1
	if j, err = m.repository.Update(j); err != nil {
		return err
	}

	return nil
}

func (m *jobManager) IncrCurrentFunction(j *Job) error {
	var err error

	j.CurrentFunction += 1
	if j, err = m.repository.Update(j); err != nil {
		return err
	}

	return nil
}
