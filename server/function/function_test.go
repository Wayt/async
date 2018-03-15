package function_test

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/wayt/async/server/function"
)

// TestCanReschedule tests Function CanReshedule
func TestCanReschedule(t *testing.T) {

	testCases := []struct {
		Function *function.Function
		Expected error
	}{
		{
			Function: &function.Function{
				RetryCount:   0,
				RetryOptions: nil,
			},
			Expected: function.ErrNoRetryOption,
		},
		{
			Function: &function.Function{
				RetryCount: 0,
				RetryOptions: &function.RetryOptions{
					RetryLimit: 1,
				},
			},
			Expected: nil,
		},
		{
			Function: &function.Function{
				RetryCount: 1,
				RetryOptions: &function.RetryOptions{
					RetryLimit: 1,
				},
			},
			Expected: function.ErrRetryLimitExceded,
		},
	}

	for _, c := range testCases {
		err := c.Function.CanReschedule()
		assert.Equal(t, err, c.Expected)
	}
}
