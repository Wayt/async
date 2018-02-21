package async

import (
	"errors"
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
	Get(uuid.UUID) (*Job, error)
	Create(*Job) (*Job, error)
	Update(*Job) (*Job, error)
	Delete(id uuid.UUID) error

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

	return j, nil
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

func (r *memoryJobRepository) Delete(id uuid.UUID) error {

	r.db.Delete(id.String())

	return nil
}

func (r *memoryJobRepository) GetPending() (*Job, bool) {

	j, ok := <-r.jobQueue
	return j, ok
}
