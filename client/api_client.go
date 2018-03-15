package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type APIClient interface {
	Info() (*V1InfoResult, error)
	Exec(*V1ExecRequest) error
}

type apiClient struct {
	httpClient *http.Client
	url        string
}

func NewAPIClient(url string, httpClient *http.Client) APIClient {
	return &apiClient{
		httpClient: httpClient,
		url:        url,
	}
}

func (c *apiClient) call(method, path string, body interface{}, result interface{}) error {

	var buf = &bytes.Buffer{}
	if body != nil {
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, c.url+PathPrefix+path, buf)
	if err != nil {
		return err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Handle http failure
	if resp.StatusCode >= 300 {
		return errors.New(resp.Status)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return err
		}
	}

	return nil
}

func (c *apiClient) get(path string, result interface{}) error {
	return c.call("GET", path, nil, result)
}

func (c *apiClient) post(path string, body, result interface{}) error {
	return c.call("POST", path, body, result)
}

func (c *apiClient) Info() (*V1InfoResult, error) {

	var result = &V1InfoResult{}
	return result, c.get("/v1/info", result)
}

func (c *apiClient) Exec(in *V1ExecRequest) error {

	return c.post("/v1/exec", in, nil)
}
