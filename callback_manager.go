package async

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type CallbackManager interface {
	GetByID(uuid.UUID) (*Callback, error)
	Create(uuid.UUID, time.Duration) (*Callback, error)
	Delete(uuid.UUID) error

	AllExpired() ([]*Callback, error)
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

func (m *callbackManager) AllExpired() ([]*Callback, error) {
	all, err := m.repository.All()
	if err != nil {
		return nil, err
	}

	callbacks := make([]*Callback, 0)
	for _, c := range all {
		if c.IsExpired() {
			callbacks = append(callbacks, c)
		}
	}

	return callbacks, nil
}

func (m *callbackManager) Delete(id uuid.UUID) error {
	return m.repository.Delete(id)
}
