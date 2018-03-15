package client

import (
	"context"
	"fmt"
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

func (d *dispatcher) dispatch(ctx context.Context, funName string, args []interface{}, data map[string]interface{}) error {
	logger.Printf("dispatcher: received execution request for %s", funName)

	fun := d.getFunc(funName)
	if fun == nil {
		return fmt.Errorf("unknown function: %s", funName)
	}

	if err := fun(ctx); err != nil {
		logger.Printf("dispatcher: %s failed: %v", funName, err)
		return err
	}

	return nil
}

func (d *dispatcher) PrintDebug() {

	logger.Printf("Having %d function(s):\n", len(d.funcs))
	for name := range d.funcs {
		logger.Printf("\t- %s\n", name)
	}
}
