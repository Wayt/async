package worker

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/wayt/async/client"
	"github.com/wayt/async/server/function"
	"github.com/wayt/async/server/job"
)

type workerState string

const (
	statePending      workerState = "pending"
	stateActive                   = "active"
	stateUnhealthy                = "unhealthy"
	stateDisconnected             = "disconnected"
)

const (
	workerRefreshInterval = 1 * time.Second
)

// Worker represents an async worker node
type Worker struct {
	sync.RWMutex

	stopCh chan struct{}

	ID          string
	Version     string
	MaxParallel int

	state workerState
	URL   string

	apiClient client.APIClient
}

func New(url string) *Worker {

	return &Worker{
		stopCh: make(chan struct{}),
		state:  statePending,
		URL:    url,
	}
}

func (w *Worker) Connect(httpClient *http.Client) error {

	w.apiClient = client.NewAPIClient(w.URL, httpClient)

	if err := w.updateInfo(); err != nil {
		return err
	}

	return nil
}

// Disconnect will stop all monitoring of the worker
// The Worker object cannot be further used without reconnecting it first.
func (w *Worker) Disconnect() {
	w.Lock()
	defer w.Unlock()

	// Resource clean up should be done only once
	if w.state == stateDisconnected {
		return
	}

	// close the chan
	close(w.stopCh)

	// TODO: Stop processing jobs

	w.state = stateDisconnected
}

func (w *Worker) updateInfo() error {

	info, err := w.apiClient.Info()
	w.checkConnectionErr(err)
	if err != nil {
		return err
	}

	w.Lock()
	defer w.Unlock()

	w.ID = info.ID
	w.Version = info.Version
	w.MaxParallel = info.MaxParallel

	return nil
}

func (w *Worker) ValidationComplete() {
	w.Lock()
	defer w.Unlock()
	if w.state != statePending {
		return
	}

	w.state = stateActive
	go w.refreshLoop()
}

// setState sets engine state
func (w *Worker) setState(state workerState) {
	w.Lock()
	defer w.Unlock()
	w.state = state
}

// IsActive returns true if the worker is active
func (w *Worker) IsActive() bool {
	w.RLock()
	defer w.RUnlock()
	return w.state == stateActive
}

// checkConnectionErr checks error from client response and adjusts worker healthy indicators
func (w *Worker) checkConnectionErr(err error) {
	if err == nil {

		if w.state == stateUnhealthy {
			w.setState(stateActive)
			log.Printf("worker: %s reconnected", w.ID)
		}
		return
	}

	if IsConnectionError(err) {
		log.Printf("worker: connection error")

		if w.state == stateActive {
			w.setState(stateUnhealthy)
			log.Printf("worker: %s unhealthy", w.ID)
		}
	}

}

// refreshLoop periodically triggers worker refresh
func (w *Worker) refreshLoop() {

	tk := time.NewTicker(workerRefreshInterval)

	for {

		// Wait for tick or if we are stopped
		select {
		case <-tk.C:
		case <-w.stopCh:
			return
		}

		w.updateInfo()
	}
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
	w.checkConnectionErr(err)

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

// IsConnectionError returns true when err is connection problem
func IsConnectionError(err error) bool {
	if strings.Contains(err.Error(), "onnection refused") {
		return true
	}
	log.Printf("IsConnectionError: %v", err)

	return false
}
