package async

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type Function func(context.Context) error

type dispatcher struct {
	sync.RWMutex
	funcs map[string]Function
}

func newDispatcher() *dispatcher {
	return &dispatcher{
		funcs: make(map[string]Function),
	}
}

func (d *dispatcher) addFunc(name string, fun Function) {
	d.Lock()
	defer d.Unlock()
	d.funcs[name] = fun
}

func (d *dispatcher) getFunc(name string) Function {
	d.RLock()
	defer d.RUnlock()
	return d.funcs[name]
}

func (d *dispatcher) Capabilities() []string {
	d.RLock()
	defer d.RUnlock()

	caps := make([]string, 0, len(d.funcs))
	for key := range d.funcs {
		caps = append(caps, key)
	}
	return caps
}

func (d *dispatcher) dispatch(ctx context.Context, funName string, args []interface{}, data map[string]interface{}) error {
	log.Printf("async: received execution request for %s", funName)

	fun := d.getFunc(funName)
	if fun == nil {
		return fmt.Errorf("unknown function: %s", funName)
	}

	if err := fun(ctx); err != nil {
		log.Printf("async: %s failed: %v", funName, err)
		return err
	}

	return nil
}

func (d *dispatcher) PrintDebug() {

	caps := d.Capabilities()
	log.Printf("async: Having %d function(s):\n", len(caps))
	for _, name := range caps {
		log.Printf("\t- %s\n", name)
	}
}
