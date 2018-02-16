package async

type Dispatcher interface {
	Dispatch(*Job) error
}

type dispatcher struct {
	manager          JobManager
	functionExecutor FunctionExecutor
}

func NewDispatcher(manager JobManager, fexec FunctionExecutor) Dispatcher {
	return &dispatcher{
		manager:          manager,
		functionExecutor: fexec,
	}
}

func (d *dispatcher) Dispatch(j *Job) error {

	var err error

	if err = d.manager.SetState(j, StateDoing); err != nil {
		return err
	}

	for i := range j.Functions {
		if i != j.CurrentFunction {
			logger.Printf("dispatcher: Skip job step %s(%d), current step is %s(%d)", j.Functions[i].Name, i, j.Functions[j.CurrentFunction].Name, j.CurrentFunction)
			continue
		}

		// Execute function
		if err = d.executeCurrentFunction(j); err != nil {
			break
		}

		if err = d.manager.IncrCurrentFunction(j); err != nil {
			break
		}
	}

	switch err {
	case errReschedule:
		return d.manager.Reschedule(j)
	case errAbort:
		return d.manager.SetState(j, StateFailed)
	case nil:
		return d.manager.SetState(j, StateDone)
	default:
		return err
	}
}

func (d *dispatcher) executeCurrentFunction(j *Job) error {
	var err error

	f := j.Functions[j.CurrentFunction]
	data := j.Data

	if err = d.manager.IncrCurrentFunctionRetryCount(j); err != nil {
		return err
	}

	if err = d.functionExecutor.Execute(*f, data); err != nil {
		if err := f.CanReschedule(); err != nil {
			logger.Printf("dispatcher: Function %s failed, cannot reschedule: %v", f.Name, err)
			return errAbort
		} else {
			logger.Printf("dispatcher: Function %s failed, rescheduling", f.Name)
			return errReschedule
		}
	}

	return nil
}
