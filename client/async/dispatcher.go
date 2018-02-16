package async

import (
	"context"
	"fmt"
	"time"
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

	fun := d.getFunc(in.Function)
	if fun == nil {
		return fmt.Errorf("unknown function: %s", in.Function)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return fun(ctx)
}

func (d *dispatcher) PrintDebug() {

	logger.Printf("Having %d function(s):\n", len(d.funcs))
	for name := range d.funcs {
		logger.Printf("\t- %s\n", name)
	}
}
