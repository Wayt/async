package server

import (
	"log"
	"net/http"

	"github.com/wayt/async/client"
	"github.com/wayt/async/server/function"
	"github.com/wayt/async/server/job"
)

const (
	WorkerStatePending = "pending"
	WorkerStateActive  = "active"
)

// Worker represents an async worker node
type Worker struct {
	ID          string
	Version     string
	MaxParallel int

	State string
	URL   string

	apiClient client.APIClient
}

func NewWorker(url string) *Worker {

	return &Worker{
		State: WorkerStatePending,
		URL:   url,
	}
}

func (w *Worker) Connect(httpClient *http.Client) error {

	w.apiClient = client.NewAPIClient(w.URL, httpClient)

	if err := w.updateInfo(); err != nil {
		return err
	}

	w.State = WorkerStateActive

	return nil
}

func (w *Worker) updateInfo() error {

	info, err := w.apiClient.Info()
	if err != nil {
		return err
	}

	w.ID = info.ID
	w.Version = info.Version
	w.MaxParallel = info.MaxParallel

	return nil
}

func (w *Worker) Process(j *job.Job) (bool, error) {

	log.Printf("worker: %s Process job: %s (%s)\n", w.ID, j.Name, j.ID.String())

	if err := w.processFunction(j.GetCurrentFunction()); err != nil {
		switch err {
		case job.ErrReschedule:
			return true, nil
		case job.ErrAbort:
			return false, nil
		default:
			return false, err
		}
	}

	return j.IncrCurrentFunction(), nil
}

func (w *Worker) processFunction(f *function.Function) error {

	f.IncrRetryCount()

	// Execute function
	err := w.apiClient.Exec(&client.V1ExecRequest{
		Function: f.Name,
		Args:     f.Args,
		Data:     nil,
	})

	if err != nil {
		log.Printf("worker: function [%s] failed: %v", f.Name, err)

		if err := f.CanReschedule(); err != nil {
			log.Printf("worker: function [%s] failed, cannot reschedule: %v", f.Name, err)
			return job.ErrAbort
		} else {
			log.Printf("worker: function [%s] failed, rescheduling", f.Name)
			return job.ErrReschedule
		}
	} else {
		log.Printf("worker: function [%s] success", f.Name)
	}

	return nil
}
