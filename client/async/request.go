package async

type FunctionRequest struct {
	Function string                 `json:"function"` // Function name
	Args     []interface{}          `json:"args"`     // Function parameters
	Data     map[string]interface{} `json:"data"`     // Job data
}

type FunctionResult struct {
	Status string `json:"status"`
	Error  error  `json:"error"`
}