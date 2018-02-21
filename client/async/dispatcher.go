package async

import (
	"context"
	"fmt"
	"net/http"
)

type Function func(context.Context) error

type dispatcher struct {
	funcs map[string]Function
}

func newDispatcher() *dispatcher {
	return &dispatcher{
		funcs: make(map[string]Function),
	}
}

func (d *dispatcher) addFunc(name string, fun Function) {
	d.funcs[name] = fun
}

func (d *dispatcher) getFunc(name string) Function {
	return d.funcs[name]
}

func (d *dispatcher) dispatch(in FunctionRequest) error {
	logger.Printf("dispatcher: received execution request for %s", in.Function)

	fun := d.getFunc(in.Function)
	if fun == nil {
		return fmt.Errorf("unknown function: %s", in.Function)
	}

	go d.WithCallback(in, fun)

	return nil

}

func (d *dispatcher) PrintDebug() {

	logger.Printf("Having %d function(s):\n", len(d.funcs))
	for name := range d.funcs {
		logger.Printf("\t- %s\n", name)
	}
}

func (d *dispatcher) WithCallback(in FunctionRequest, fun Function) {

	ctx := context.Background()
	ctx, cancel := context.WithDeadline(ctx, in.Callback.ExpiredAt)
	defer cancel()

	statusCode := http.StatusOK

	if err := fun(ctx); err != nil {
		logger.Printf("dispatcher: %s failed: %v", in.Function, err)

		statusCode = http.StatusInternalServerError
	}

	if err := in.Callback.Call(statusCode); err != nil {
		logger.Printf("dispatcher: failed to call Callback %s: %v", in.Callback.ID, err)
	}
}
