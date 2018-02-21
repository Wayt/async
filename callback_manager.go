package async

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type CallbackManager interface {
	GetByID(uuid.UUID) (*Callback, error)
	Create(uuid.UUID, time.Duration) (*Callback, error)
}

type callbackManager struct {
	repository CallbackRepository
}

func NewCallbackManager(repo CallbackRepository) CallbackManager {

	return &callbackManager{
		repository: repo,
	}
}

func (m *callbackManager) GetByID(id uuid.UUID) (*Callback, error) {
	return m.repository.Get(id)
}

func (m *callbackManager) Create(jobID uuid.UUID, expiration time.Duration) (*Callback, error) {

	c := &Callback{
		JobID:     jobID,
		ExpiredAt: time.Now().Add(expiration),
	}

	var err error
	if c, err = m.repository.Create(c); err != nil {
		return nil, err
	}

	return c, nil
}
