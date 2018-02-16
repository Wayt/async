package async

type Poller interface {
	Poll(Dispatcher)
}

type poller struct {
	repository JobRepository
}

func NewPoller(repo JobRepository) Poller {

	return &poller{
		repository: repo,
	}
}

func (p *poller) Poll(dsp Dispatcher) {
	for {

		j, ok := p.repository.GetPending()
		if !ok {
			logger.Printf("poller: received a stop signal from repository")
			return
		}

		logger.Printf("poller: received: %s - %s", j.ID, j.Name)

		if err := dsp.Dispatch(j); err != nil {
			logger.Printf("poller: error during dispatch: %v", err)
		}
	}
}
