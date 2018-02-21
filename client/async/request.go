package async

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type FunctionCallback struct {
	ID        uuid.UUID `json:"callback_id"`
	URL       string    `json:"url"`
	ExpiredAt time.Time `json:"expired_at"`
}

type FunctionRequest struct {
	Function string                 `json:"function"` // Function name
	Args     []interface{}          `json:"args"`     // Function parameters
	Data     map[string]interface{} `json:"data"`     // Job data
	Callback FunctionCallback       `json:"callback"` // Callback information
}

type FunctionResult struct {
	Status string `json:"status"`
	Error  error  `json:"error"`
}
