package server

import (
	"encoding/json"
	"net/http"

	"github.com/wayt/async/server/function"
	"github.com/wayt/async/server/worker"
)

type handler func(c *context, w http.ResponseWriter, r *http.Request)

var routes = map[string]map[string]handler{
	"POST": {
		"/v1/worker": postWorker,
		"/v1/job":    postJob,
	},
	"GET": {
		"/v1/worker": getWorkers,
	},
}

type context struct {
	server *Server
}

func setupHandlers(s *Server) {

	context := &context{
		server: s,
	}

	for method, mappings := range routes {
		for route, fct := range mappings {

			localMethod := method
			localRoute := route
			localFct := fct

			wrap := func(w http.ResponseWriter, r *http.Request) {
				localFct(context, w, r)
			}
			s.router.Path(localRoute).Methods(localMethod).HandlerFunc(wrap)
		}
	}

}

func postJob(c *context, w http.ResponseWriter, r *http.Request) {
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

// TODO: hearthbeat
func postWorker(c *context, w http.ResponseWriter, r *http.Request) {

	var in struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.server.addWorker(in.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getWorkers(c *context, w http.ResponseWriter, r *http.Request) {

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
