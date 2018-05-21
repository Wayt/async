package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"

	uuid "github.com/satori/go.uuid"
	pb "github.com/wayt/async/pb/server"
	"github.com/wayt/async/server/broker"
	"github.com/wayt/async/server/function"
	"github.com/wayt/async/server/job"
	"github.com/wayt/async/server/worker"
	"google.golang.org/grpc"

	"github.com/spf13/viper"
)

var config = viper.New()

func init() {
	config.SetEnvPrefix("async_server")
	config.SetDefault("bind", ":8080")
	config.SetDefault("http", ":8000")

	config.AutomaticEnv()
}

type Server struct {
	sync.RWMutex

	workers        map[string]*worker.Worker // Registered workers
	pendingWorkers map[string]*worker.Worker // Registration pending workers

	broker broker.Broker
	router *mux.Router

	gRPCServer *grpc.Server
}

func New() *Server {

	s := &Server{
		workers:        make(map[string]*worker.Worker),
		pendingWorkers: make(map[string]*worker.Worker),
		broker:         broker.NewMemoryBroker(),
		gRPCServer:     grpc.NewServer(),
	}

	pb.RegisterServerServer(s.gRPCServer, s)
	setupHandlers(s)

	return s
}

func (s *Server) Run() error {

	go func() { log.Fatal(http.ListenAndServe(config.GetString("http"), s.router)) }()

	bind := config.GetString("bind")

	lis, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}

	log.Printf("server: Serving gRPC API on %s", bind)
	if err := s.gRPCServer.Serve(lis); err != nil {
		return err
	}

	return nil
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

func (s *Server) RegisterWorker(ctx context.Context, in *pb.RegisterWorkerRequest) (*pb.RegisterWorkerReply, error) {
	if in.Address == "" {
		return nil, errors.New("missing address")
	}

	log.Printf("server: new worker registration request for %s", in.Address)

	w := worker.New(in.Address)

	s.Lock()
	s.pendingWorkers[in.Address] = w
	s.Unlock()

	go s.connectWorker(w)

	return &pb.RegisterWorkerReply{
		State: string(w.State),
	}, nil
}

func (s *Server) connectWorker(w *worker.Worker) {

	err := w.Connect()

	s.Lock()
	defer s.Unlock()

	if err != nil {
		log.Printf("server: failed to validate pending worker %s: %v", w.Address, err)
		delete(s.pendingWorkers, w.Address)
		return
	}

	// Make sure worker ID is unique
	if old, exists := s.workers[w.ID]; exists {
		// TODO: better handle id conflict
		log.Printf("server: worker ID duplicated, %s shared by %s and %s. Removing older worker.", w.ID, old.Address, w.Address)
		old.Disconnect()
		delete(s.workers, old.ID)
	}

	// Engine validated, move from pendingWorkers
	delete(s.pendingWorkers, w.Address)
	w.ValidationComplete()
	s.workers[w.ID] = w

	go func() {
		<-w.Stopped()

		log.Printf("server: removing worker %s", w.ID)

		s.Lock()
		defer s.Unlock()
		delete(s.workers, w.ID)
	}()

	// Start sending job to worker
	// FIXME: this should be done in the worker
	// So parallel processing can be updated on the fly
	// And processing can be stopped when worker stopserver/handlers.go
	var i int32
	for i = 0; i < w.MaxParallel; i++ {
		go s.broker.Consume(w)
	}

	log.Printf("server: registered Worker %s at %s with max_parallel %d and capabilities: %v", w.ID, w.Address, w.MaxParallel, w.Capabilities)
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
