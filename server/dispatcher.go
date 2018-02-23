package server

type Dispatcher interface {
	Dispatch(*Job) error
}

type dispatcher struct {
	jobMgr           JobManager
	callbackMgr      CallbackManager
	functionExecutor FunctionExecutor
}

func NewDispatcher(jobMgr JobManager, callbackMgr CallbackManager, fexec FunctionExecutor) Dispatcher {
	return &dispatcher{
		jobMgr:           jobMgr,
		callbackMgr:      callbackMgr,
		functionExecutor: fexec,
	}
}

func (d *dispatcher) Dispatch(j *Job) error {

	var err error

	if err = d.jobMgr.SetState(j, StateDoing); err != nil {
		return err
	}

	// Execute function
	err = d.executeCurrentFunction(j)

	switch err {
	case errReschedule:
		return d.jobMgr.Reschedule(j)
	case errAbort:
		return d.jobMgr.SetState(j, StateFailed)
	case nil:
		// return d.jobMgr.SetState(j, StateDone)
		return nil
	default:
		return err
	}
}

func (d *dispatcher) executeCurrentFunction(j *Job) error {
	var err error

	f := j.Functions[j.CurrentFunction]
	data := j.Data

	// Create a callback
	callback, err := d.callbackMgr.Create(j.ID, DefaultCallbackTimeout)
	if err != nil {
		return err
	}

	if err = d.jobMgr.IncrCurrentFunctionRetryCount(j); err != nil {
		return err
	}

	if err = d.functionExecutor.Execute(callback, *f, data); err != nil {
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
