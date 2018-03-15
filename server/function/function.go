package function

import "errors"

var (
	// ErrNoRetryOption is returned by CanReschedule when a Function has no RetryOptions
	ErrNoRetryOption = errors.New("no retry option")

	// ErrRetryLimitExceded is returned by CanReschedule when a Function has already used all it retries
	ErrRetryLimitExceded = errors.New("retry limit exceded")
)

// RetryOptions define retry policy for a given Function
type RetryOptions struct {
	RetryLimit int32 `json:"retry_limit"`
}

// Function represents a Job function
type Function struct {
	Name string        `json:"name"`
	Args []interface{} `json:"args,omitempty"`
	//TODO: Delay time.Duration // To Delay a task
	RetryCount   int32         `json:"retry_count"`
	RetryOptions *RetryOptions `json:"retry_options,omitempty"`
}

// CanReschedule returns an error if this function cannot be rescheduled
// See returned error for exact reason
func (f *Function) CanReschedule() error {
	if f.RetryOptions == nil {
		return ErrNoRetryOption
	}

	if f.RetryCount >= f.RetryOptions.RetryLimit {
		return ErrRetryLimitExceded
	}

	return nil
}

func (f *Function) IncrRetryCount() {

	f.RetryCount += 1
}
