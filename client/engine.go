package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

var (
	config        = viper.New()
	DefaultEngine *Engine

	ErrCannotConnect = errors.New("cannot connect to server")
)

func init() {
	config.SetEnvPrefix("async")
	hostname, _ := os.Hostname()

	config.SetDefault("id", hostname)
	config.SetDefault("bind", ":8179")
	config.SetDefault("advertise_url", "http://127.0.0.1:8179")
	config.SetDefault("server_url", "http://127.0.0.1:8080")
	config.SetDefault("worker", 2)

	config.AutomaticEnv()

	DefaultEngine = NewEngine(
		config.GetString("id"),
		config.GetString("bind"),
		config.GetString("advertise_url"),
		config.GetString("server_url"),
		config.GetInt("worker"))
}

const (
	Version    = "v0.0.0 -- HEAD"
	PathPrefix = "/_/async"
)

type Engine struct {
	id           string
	advertiseURL string
	serverURL    string
	server       *http.Server
	httpClient   *http.Client
	dispatcher   *dispatcher
	workerCount  int
}

func NewEngine(id, bind, advertiseURL, serverURL string, workerCount int) *Engine {

	d := newDispatcher()
	e := &Engine{
		id:           id,
		advertiseURL: advertiseURL,
		serverURL:    serverURL,
		dispatcher:   d,
		httpClient:   http.DefaultClient,
		workerCount:  workerCount,
	}

	e.initServer(bind)

	return e
}

func (e *Engine) initServer(bind string) {

	mux := mux.NewRouter()
	mux.Methods("POST").Path(PathPrefix + "/v1/exec").HandlerFunc(e.postV1Exec)
	mux.Methods("GET").Path(PathPrefix + "/v1/info").HandlerFunc(e.getV1Info)

	e.server = &http.Server{
		Addr:           bind,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func (e *Engine) SetHTTPClient(httpClient *http.Client) {
	e.httpClient = httpClient
}

func Func(name string, fun Function) { DefaultEngine.Func(name, fun) }
func (e *Engine) Func(name string, fun Function) {
	e.dispatcher.addFunc(name, fun)
}

func Run() error { return DefaultEngine.Run() }
func (e *Engine) Run() error {

	e.dispatcher.PrintDebug()

	logger.Printf("Listening and serving HTTP on %s\n", e.server.Addr)

	go e.server.ListenAndServe()

	if err := e.connectWithRetry(); err != nil {
		e.server.Shutdown(context.Background())
		return err
	}

	// TODO: start heartbeat
	for {
		time.Sleep(1)
	}

	return nil
}

type V1ExecRequest struct {
	Function string                 `json:"function"` // Function name
	Args     []interface{}          `json:"args"`     // Function parameters
	Data     map[string]interface{} `json:"data"`     // Job data
}

func (e *Engine) postV1Exec(w http.ResponseWriter, r *http.Request) {

	logger.Printf("handler: %s - %s %s\n", r.RemoteAddr, r.Method, r.URL.Path)

	var in V1ExecRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		logger.Printf("handler: failed decoding request payload: %v", err)
		http.Error(w, "failed decoding request payload", http.StatusBadRequest)
		return
	}

	if err := e.dispatcher.dispatch(r.Context(), in.Function, in.Args, in.Data); err != nil {
		logger.Printf("handler: failed to dispatch %s: %v", in.Function, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type V1InfoResult struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	MaxParallel int    `json:"max_parallel"`
}

func (e *Engine) getV1Info(w http.ResponseWriter, r *http.Request) {

	logger.Printf("HTTP Info request from %s", r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(V1InfoResult{
		ID:          e.id,
		Version:     Version,
		MaxParallel: e.workerCount,
	}); err != nil {
		logger.Printf("getV1Info: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (e *Engine) connectWithRetry() error {
	// TODO: backoff
	for i := 0; i <= 5; i++ {

		if err := e.connect(); err != nil {
			return err
		} else {
			// Connected !
			return nil
		}
	}

	return ErrCannotConnect
}

func (e *Engine) connect() error {

	logger.Printf("connect: attempting on %s", e.serverURL)

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(struct {
		URL string `json:"url"`
	}{
		URL: e.advertiseURL,
	}); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", e.serverURL+"/v1/worker", buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Printf("connect: reponse from server: %s", resp.Status)
		return ErrCannotConnect
	}

	return nil
}
