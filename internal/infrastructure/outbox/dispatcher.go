package outbox

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/rcarvalho-pb/payment_system-go/internal/application/worker"
	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
)

type Dispatcher struct {
	Repo         Repository
	EventBus     worker.EventPublisher
	PollInterval time.Duration
	BatchSize    int
}

func (d *Dispatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(d.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Println("ticker")
			d.DispatchOnce()
		}
	}
}

func (d *Dispatcher) DispatchOnce() {
	log.Println("Dispatch one")
	events, err := d.Repo.FindUnpublished(d.BatchSize)
	if err != nil {
		log.Println(err.Error())
		return // logável, mas não fatal
	}

	log.Println(events)
	log.Println("events")

	for _, evt := range events {
		var payload any

		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			continue
		}

		domainEvent := event.Event{
			Type:    evt.Type,
			Payload: payload,
		}

		if err := d.EventBus.Publish(domainEvent); err != nil {
			continue
		}

		_ = d.Repo.MarkPublished(evt.ID)
	}
}
