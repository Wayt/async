package async

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

func (f *Function) CanReschedule() bool {
	if f.RetryOptions == nil {
		return false
	}

	if f.RetryCount < f.RetryOptions.RetryLimit {
		return true
	}

	return false
}
