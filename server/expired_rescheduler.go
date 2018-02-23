package server

import (
	"net/http"
	"time"
)

type ExpiredRescheduler struct {
	jm JobManager
	cm CallbackManager
}

func NewExpiredRescheduler(jm JobManager, cm CallbackManager) *ExpiredRescheduler {
	return &ExpiredRescheduler{
		jm: jm,
		cm: cm,
	}
}

func (r *ExpiredRescheduler) Run() {
	for {
		time.Sleep(1 * time.Second)

		// Get all expired callback
		expiredList, err := r.cm.AllExpired()
		if err != nil {
			logger.Printf("expired_rescheduler: fail to get expired callback: %v", err)
			continue
		}

		if len(expiredList) > 0 {
			logger.Printf("expired_rescheduler: found %d expired callback", len(expiredList))
			for _, c := range expiredList {
				logger.Printf("expired_rescheduler: doing callback %s", c.ID)
				if err := r.jm.HandleCallback(c, http.StatusRequestTimeout); err != nil {
					logger.Printf("expired_rescheduler: fail to handle callback %s: %v", c.ID, err)
					continue
				}

				if err := r.cm.Delete(c.ID); err != nil {
					logger.Printf("expired_rescheduler: fail to delete callback %s: %v", c.ID, err)
					continue
				}
			}
		}
	}
}
