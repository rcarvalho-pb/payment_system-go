package outbox

import (
	"context"
	"encoding/json"
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
			d.DispatchOnce()
		}
	}
}

func (d *Dispatcher) DispatchOnce() {
	events, err := d.Repo.FindUnpublished(d.BatchSize)
	if err != nil {
		return // logável, mas não fatal
	}

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
