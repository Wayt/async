package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"

	"github.com/wayt/async/server/function"
	"github.com/wayt/async/server/job"
	"github.com/wayt/async/server/worker"
)

type handler func(c *handlerContext, w http.ResponseWriter, r *http.Request)

var routes = map[string]map[string]handler{
	"POST": {
		"/v1/job": postJob,
	},
	"GET": {
		"/v1/worker":       getWorkers,
		"/v1/job":          getJobs,
		"/v1/job/{job_id}": getJob,
	},
}

type handlerContext struct {
	server *Server
}

func setupHandlers(s *Server) {

	s.router = mux.NewRouter()

	c := &handlerContext{
		server: s,
	}

	for method, mappings := range routes {
		for route, fct := range mappings {

			localMethod := method
			localRoute := route
			localFct := fct

			wrap := func(w http.ResponseWriter, r *http.Request) {
				localFct(c, w, r)
			}
			s.router.Path(localRoute).Methods(localMethod).HandlerFunc(wrap)
		}
	}

}

func postJob(c *handlerContext, w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name      string                 `json:"name" binding:"required"`
		Functions []*function.Function   `json:"functions" binding:"required"`
		Data      map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	j, err := c.server.CreateJob(in.Name, in.Functions, in.Data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.NewEncoder(w).Encode(j); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getWorkers(c *handlerContext, w http.ResponseWriter, r *http.Request) {

	workers := c.server.listWorkers()

	result := struct {
		Workers []*worker.Worker
	}{
		Workers: workers,
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getJobs(c *handlerContext, w http.ResponseWriter, r *http.Request) {
	jobs := c.server.broker.List()

	result := struct {
		Count int
		Jobs  []*job.Job
	}{
		Count: len(jobs),
		Jobs:  jobs,
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getJob(c *handlerContext, w http.ResponseWriter, r *http.Request) {

	jobIDStr := mux.Vars(r)["job_id"]

	jobID, err := uuid.FromString(jobIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	j, err := c.server.broker.Get(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	result := struct {
		Job *job.Job
	}{
		Job: j,
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
