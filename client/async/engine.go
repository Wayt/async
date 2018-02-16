package async

import (
	"fmt"
	"net/http"
	"time"
)

const (
	Port = 8179
	Path = "/_/async/v1/exec"
)

type Engine struct {
	server     *http.Server
	dispatcher *dispatcher
}

func NewEngine() *Engine {

	d := newDispatcher()

	mux := http.NewServeMux()
	mux.Handle(Path, &handler{dispatcher: d})

	return &Engine{
		server: &http.Server{
			Addr:           fmt.Sprintf(":%d", Port),
			Handler:        mux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		dispatcher: d,
	}
}

func (e *Engine) Func(name string, fun Function) {
	e.dispatcher.addFunc(name, fun)
}

func (e *Engine) Run() error {

	e.dispatcher.PrintDebug()

	logger.Printf("Listening and serving HTTP on %s\n", e.server.Addr)

	return e.server.ListenAndServe()
}
