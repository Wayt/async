package server

import (
	"fmt"
	"net/http"

	"github.com/satori/go.uuid"
)

type JobManager interface {
	GetByID(uuid.UUID) (*Job, error)
	Create(string, []*Function, map[string]interface{}) (*Job, error)

	SetState(*Job, string) error
	Reschedule(*Job) error
	RescheduleID(uuid.UUID) error
	IncrCurrentFunctionRetryCount(*Job) error
	IncrCurrentFunction(*Job) error

	HandleCallback(*Callback, int) error
}

type jobManager struct {
	repository JobRepository
}

func NewJobManager(repo JobRepository) JobManager {

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

	var err error
	if j, err = m.repository.Create(j); err != nil {
		return nil, err
	}

	return j, m.repository.Schedule(j)
}

func (m *jobManager) SetState(j *Job, state string) error {
	var err error

	j.State = state
	if j, err = m.repository.Update(j); err != nil {
		return err
	}

	return nil
}

func (m *jobManager) RescheduleID(id uuid.UUID) error {

	j, err := m.GetByID(id)
	if err != nil {
		return err
	}

	return m.Reschedule(j)
}

func (m *jobManager) Reschedule(j *Job) error {

	logger.Printf("job_manager: Reschedule job %s - %s", j.ID, j.Name)

	currentFunc := j.GetCurrentFunction()
	if err := currentFunc.CanReschedule(); err != nil {
		logger.Printf("job_manager: Cannot reschedule %s - %s: %v", j.ID, j.Name, err)

		if err := m.SetState(j, StateFailed); err != nil {
			return err
		}
		return nil
	} else {
		if err := m.SetState(j, StatePending); err != nil {
			return err
		}

		return m.repository.Schedule(j)
	}
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

func (m *jobManager) HandleCallback(c *Callback, statusCode int) error {

	job, err := m.GetByID(c.JobID)
	if err != nil {
		return err
	}

	logger.Printf("job_manager: handle callback %s for job %s, statusCode: %d", c.ID, job.ID, statusCode)

	switch statusCode {
	case http.StatusInternalServerError:
		return m.Reschedule(job)
	case http.StatusRequestTimeout:
		return m.Reschedule(job)
	case http.StatusOK:
		if err = m.IncrCurrentFunction(job); err != nil {
			return err
		}
		return m.Reschedule(job)
	default:
		err = fmt.Errorf("unknown StatusCode %d", statusCode)
		logger.Printf("job_manager: %v in Callback %s", err, c.ID)
		return err
	}
}