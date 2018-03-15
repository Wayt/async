package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/satori/go.uuid"
)

type FunctionCallback struct {
	ID        uuid.UUID `json:"callback_id"`
	URL       string    `json:"url"`
	ExpiredAt time.Time `json:"expired_at"`
}

func (c *FunctionCallback) Call(statusCode int) error {

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(map[string]interface{}{
		"status_code": statusCode,
	}); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.URL, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	// TODO: retry based on resp.StatusCode

	return resp.Body.Close()
}
