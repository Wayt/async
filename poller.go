package async

import "log"

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
			log.Printf("Poller received a stop from repository")
			return
		}

		log.Printf("Polled %s - %s", j.ID, j.Name)

		if err := dsp.Dispatch(j); err != nil {
			log.Printf("error during dispatch: %v", err)
		}
	}
}
