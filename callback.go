package async

import (
	"fmt"
	"time"

	cache "github.com/pmylund/go-cache"
	uuid "github.com/satori/go.uuid"
	cli "github.com/wayt/async/client/async"
)

type Callback struct {
	ID        uuid.UUID `json:"callback_id"`
	JobID     uuid.UUID `json:"job_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func (c *Callback) URL() string {

	// FIXME: this should be configurable or automatic
	apiURL := "http://127.0.0.1:8080"

	return fmt.Sprintf("%s/v1/job/%s/callback/%s", apiURL, c.JobID.String(), c.ID.String())
}

func (c *Callback) BuildFunctionCallback() cli.FunctionCallback {

	return cli.FunctionCallback{
		ID:        c.ID,
		URL:       c.URL(),
		ExpiredAt: c.ExpiredAt,
	}
}

// CallbackRepository

type CallbackRepository interface {
	Get(uuid.UUID) (*Callback, error)
	Create(*Callback) (*Callback, error)
	Delete(uuid.UUID) error
}

// In memory callback repository

type callbackRepository struct {
	db *cache.Cache
}

func NewCallbackRepository() CallbackRepository {
	return &callbackRepository{
		db: cache.New(cache.NoExpiration, time.Minute),
	}
}

func (r *callbackRepository) Get(id uuid.UUID) (*Callback, error) {

	i, ok := r.db.Get(id.String())
	if !ok {
		return nil, ErrNotFound
	}

	return i.(*Callback), nil
}

func (r *callbackRepository) Create(c *Callback) (*Callback, error) {

	var err error
	if c.ID, err = uuid.NewV4(); err != nil {
		return nil, err
	}

	c.CreatedAt = time.Now()

	r.db.Set(c.ID.String(), c, cache.NoExpiration)

	return c, nil
}

func (r *callbackRepository) Delete(id uuid.UUID) error {

	r.db.Delete(id.String())

	return nil
}
