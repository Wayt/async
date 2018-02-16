package async

import "errors"

var (
	ErrNoRetryOption     = errors.New("no retry option")
	ErrRetryLimitExceded = errors.New("retry limit exceded")
)

type RetryOptions struct {
	RetryLimit int32 `json:"retry_limit"`
}

type Function struct {
	Name string        `json:"name"`
	Args []interface{} `json:"args,omitempty"`
	//TODO: Delay time.Duration // To Delay a task
	RetryCount   int32         `json:"retry_count"`
	RetryOptions *RetryOptions `json:"retry_options,omitempty"`
}

func (f *Function) CanReschedule() error {
	if f.RetryOptions == nil {
		return ErrNoRetryOption
	}

	if f.RetryCount >= f.RetryOptions.RetryLimit {
		return ErrRetryLimitExceded
	}

	return nil
}
