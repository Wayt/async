package job

import (
	"errors"
	"time"

	"github.com/satori/go.uuid"
	"github.com/wayt/async/server/function"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrReschedule = errors.New("reschedule")
	ErrAbort      = errors.New("abort")
)

type Job struct {
	ID              uuid.UUID              `json:"job_id"`
	Name            string                 `json:"name"`
	Functions       []*function.Function   `json:"functions"`
	CurrentFunction int                    `json:"current_function"`
	Data            map[string]interface{} `json:"data"`
	CreatedAt       time.Time              `json:"created_at"`
	ScheduledAt     time.Time              `json:"scheduled_at"`
}

func (j *Job) GetCurrentFunction() *function.Function {
	return j.Functions[j.CurrentFunction]
}

func (j *Job) IncrCurrentFunction() bool {
	if j.CurrentFunction == len(j.Functions)-1 {
		return false
	}

	j.CurrentFunction += 1
	return true
}
