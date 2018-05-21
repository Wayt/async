package worker

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	pb "github.com/wayt/async/pb/worker"
	"github.com/wayt/async/server/function"
	"github.com/wayt/async/server/job"
	"google.golang.org/grpc"
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
	maxConnectionFailure  = 3
)

// Worker represents an async worker node
type Worker struct {
	sync.RWMutex

	stopCh chan struct{}

	ID           string
	Version      string
	MaxParallel  int32
	Capabilities []string // List of function handled by worker

	ConnectionFailure int32
	State             workerState
	Address           string

	client pb.WorkerClient
}

func New(address string) *Worker {

	return &Worker{
		stopCh:            make(chan struct{}),
		ConnectionFailure: 0,
		State:             statePending,
		Address:           address,
	}
}

func (w *Worker) GetCapabilities() []string { w.RLock(); defer w.RUnlock(); return w.Capabilities }

func (w *Worker) Connect() error {

	log.Printf("worker: connect %s", w.ID)

	conn, err := grpc.Dial(w.Address, grpc.WithInsecure(),
		grpc.WithBackoffConfig(grpc.BackoffConfig{
			MaxDelay: time.Second * 10,
		}))
	if err != nil {
		return err
	}

	go func() {
		tk := time.NewTicker(1 * time.Second)
		for range tk.C {
			if !w.IsActive() {
				conn.Close()
				return
			}
		}
	}()

	w.client = pb.NewWorkerClient(conn)

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

	log.Printf("worker: disconnecting %s", w.ID)

	// Resource clean up should be done only once
	if w.State == stateDisconnected {
		return
	}

	// close the chan
	close(w.stopCh)

	// TODO: Stop processing jobs
	// Currently, the broker stop sending job when worker is stopped

	w.State = stateDisconnected

	w.client = nil
}

func (w *Worker) Stopped() <-chan struct{} {
	return w.stopCh
}

func (w *Worker) updateInfo() error {

	info, err := w.client.Info(context.Background(), &pb.InfoRequest{})
	w.checkConnectionErr(err)
	if err != nil {
		return err
	}

	w.Lock()
	defer w.Unlock()

	w.ID = info.GetId()
	w.Version = info.GetVersion()
	w.MaxParallel = info.GetMaxParallel()
	w.Capabilities = info.GetCapabilities()

	return nil
}

func (w *Worker) ValidationComplete() {
	w.Lock()
	defer w.Unlock()
	if w.State != statePending {
		return
	}

	w.State = stateActive
	go w.refreshLoop()
}

// setState sets engine state
func (w *Worker) setState(state workerState) {
	w.Lock()
	defer w.Unlock()
	w.State = state
}

// IsActive returns true if the worker is active
func (w *Worker) IsActive() bool {
	w.RLock()
	defer w.RUnlock()
	return w.State == stateActive
}

// checkConnectionErr checks error from client response and adjusts worker healthy indicators
func (w *Worker) checkConnectionErr(err error) {
	if err == nil {

		if w.State == stateUnhealthy {
			w.setState(stateActive)
			log.Printf("worker: %s reconnected", w.ID)
		}
		return
	}

	if IsConnectionError(err) {
		log.Printf("worker: connection error")
		w.Lock()
		w.ConnectionFailure += 1
		w.Unlock()

		if w.State == stateActive {
			w.setState(stateUnhealthy)
			log.Printf("worker: %s unhealthy", w.ID)
		}

		if w.ConnectionFailure >= maxConnectionFailure {
			w.Disconnect()

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

		if w.IsActive() {
			w.updateInfo()
		} else {
			if err := w.Connect(); err != nil {
				log.Printf("worker: %s", err)
			}
		}
	}
}

func (w *Worker) Process(j *job.Job) (bool, error) {

	log.Printf("worker: %s Process job: %s (%s)\n", w.ID, j.Name, j.ID)

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

	_, err := w.client.Exec(context.Background(), &pb.ExecRequest{
		Function: f.Name,
		// Args:     args,
		// Data:     nil,
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
	if err == grpc.ErrClientConnClosing ||
		err == grpc.ErrClientConnTimeout {
		return true
	}

	if strings.Contains(err.Error(), "onnection refused") {
		return true
	}
	log.Printf("IsConnectionError: %v", err)

	return false
}
