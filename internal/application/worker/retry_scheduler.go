package worker

import (
	"time"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
)

type RetryScheduler struct {
	EventBus  EventPublisher
	MaxRetry  int
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

func (r *RetryScheduler) ScheduleRetry(payload event.PaymentRequestPayload) {
	if payload.Attempt >= r.MaxRetry {
		return
	}

	delay := min(r.BaseDelay*time.Duration(1<<(payload.Attempt-1)), r.MaxDelay)

	nextPayload := event.PaymentRequestPayload{
		InvoiceID: payload.InvoiceID,
		Amount:    payload.Amount,
		Attempt:   payload.Attempt + 1,
	}

	go func() {
		time.Sleep(delay)
		r.EventBus.Publish(event.Event{
			Type:    event.PaymentRequested,
			Payload: nextPayload,
		})
	}()
}
