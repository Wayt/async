package async

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/pkg/errors"
	cli "github.com/wayt/async/client/async"
)

type FunctionExecutor interface {
	Execute(Function, map[string]interface{}) error
}

type defaultFunctionExecutor struct {
	url        string
	HTTPClient *http.Client
}

func NewFunctionExecutor(url string) FunctionExecutor {

	return &defaultFunctionExecutor{
		url:        url,
		HTTPClient: http.DefaultClient,
	}
}

func (e *defaultFunctionExecutor) Execute(f Function, data map[string]interface{}) error {

	body, _ := json.Marshal(cli.FunctionRequest{
		Function: f.Name,
		Args:     f.Args,
		Data:     data,
	})

	req, err := http.NewRequest("POST", e.url+cli.Path, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "fail to create Request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "fail to execute query")
	}
	// Nothing to read
	resp.Body.Close()

	log.Printf("[%s] - %d - %s %s", f.Name, resp.StatusCode, req.Method, req.URL.String())

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return errors.New("function failed")
	}

	log.Printf("%s executed", f.Name)

	return nil
}
