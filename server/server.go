package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"github.com/wayt/async/server/broker"
	"github.com/wayt/async/server/function"
	"github.com/wayt/async/server/job"
	"github.com/wayt/async/server/worker"

	"github.com/spf13/viper"
)

var config = viper.New()

func init() {
	config.SetEnvPrefix("async_server")
	config.SetDefault("bind", ":8080")
	config.SetDefault("executor_url", "http://127.0.0.1:8179")
	config.SetDefault("callback_url", "http://127.0.0.1:8080")

	config.AutomaticEnv()
}

type Server struct {
	sync.RWMutex

	router         *mux.Router
	workers        map[string]*worker.Worker // Registered workers
	pendingWorkers map[string]*worker.Worker // Registration pending workers

	broker broker.Broker

	httpClient *http.Client
}

func New() *Server {

	// cr := NewCallbackRepository()
	// cm := NewCallbackManager(cr, viper.GetString("callback_url"))

	// d := NewDispatcher(jm, cm, NewFunctionExecutor(viper.GetString("executor_ur")))
	// p := NewPoller(jr)
	// go p.Poll(d)

	// rescheduler := NewExpiredRescheduler(jm, cm)
	// go rescheduler.Run()

	s := &Server{
		router:         mux.NewRouter(),
		workers:        make(map[string]*worker.Worker),
		pendingWorkers: make(map[string]*worker.Worker),
		broker:         broker.NewMemoryBroker(),
		httpClient:     http.DefaultClient,
	}

	// NewHttpJobHandler(e, jm)
	// 	NewHttpCallbackHandler(e, cm, jm)

	setupHandlers(s)

	return s
}

func (s *Server) Run() {
	srv := &http.Server{
		Handler: s.router,
		Addr:    config.GetString("bind"),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("server: Serving HTTP API on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func (s *Server) CreateJob(name string, functions []*function.Function, data map[string]interface{}) (*job.Job, error) {

	if len(functions) == 0 {
		return nil, fmt.Errorf("cannot create a job with empty functions")
	}

	j := &job.Job{
		ID:              uuid.NewV4(),
		Name:            name,
		Functions:       functions,
		Data:            data,
		CurrentFunction: 0,
		CreatedAt:       time.Now(),
	}

	if err := s.broker.Schedule(j); err != nil {
		return nil, err
	}

	log.Printf("server: received job [%s] with id [%s]", j.Name, j.ID)
	return j, nil
}

func (s *Server) addWorker(url string) error {

	log.Printf("server: new worker registration request for %s", url)

	w := worker.New(url)

	s.Lock()
	s.pendingWorkers[url] = w
	s.Unlock()

	go s.connectWorker(w)

	return nil
}

func (s *Server) connectWorker(w *worker.Worker) {

	if err := w.Connect(s.httpClient); err != nil {
		log.Printf("server: failed to validate pending worker %s: %v", w.URL, err)
		return
	}

	s.Lock()
	defer s.Unlock()

	// Only validate workers from pendingWorkers list
	if _, exists := s.pendingWorkers[w.URL]; !exists {
		return
	}

	// Make sure worker ID is unique
	if old, exists := s.workers[w.ID]; exists {
		// TODO: better handle id conflict
		log.Printf("server: worker ID duplicated, %s shared by %s and %s. Removing older worker.", w.ID, old.URL, w.URL)
		old.Disconnect()
		delete(s.workers, old.ID)
	}

	// Engine validated, move from pendingWorkers
	delete(s.pendingWorkers, w.URL)
	w.ValidationComplete()
	s.workers[w.ID] = w

	// Start sending job to worker
	// FIXME: this should be done in the worker
	// So parallel processing can be updated on the fly
	// And processing can be stopped when worker stop
	for i := 0; i < w.MaxParallel; i++ {
		go s.broker.Consume(w)
	}

	log.Printf("server: registered Worker %s at %s with max_parallel %d", w.ID, w.URL, w.MaxParallel)
}

// listWorkers returns all the workers in the server
// This is for reporting, not scheduling, hence pendingWorkers are included
func (s *Server) listWorkers() []*worker.Worker {
	s.RLock()
	defer s.RUnlock()

	list := make([]*worker.Worker, 0, len(s.pendingWorkers)+len(s.workers))
	for _, w := range s.pendingWorkers {
		list = append(list, w)
	}
	for _, w := range s.workers {
		list = append(list, w)
	}

	return list
}

// ListActiveWorkers returns all validated workers in the server
func (s *Server) listActiveWorkers() []*worker.Worker {
	s.RLock()
	defer s.RUnlock()

	list := make([]*worker.Worker, 0, len(s.workers))
	for _, w := range s.workers {
		list = append(list, w)
	}

	return list
}
