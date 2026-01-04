package worker

import (
	"errors"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_system-go/internal/domain/payment"
	"github.com/rcarvalho-pb/payment_system-go/internal/infra/logging"
	"github.com/rcarvalho-pb/payment_system-go/internal/infra/metrics"
)

type PaymentProcessor struct {
	Repo     payment.Repository
	EventBus EventPublisher
	Retry    Scheduler
	Logger   logging.Logger
	Metrics  *metrics.Counters
	Executor PaymentExecutor
}

type EventPublisher interface {
	Publish(event.Event) error
}

func (p *PaymentProcessor) Handle(evt event.Event) error {
	if evt.Type != event.PaymentRequested {
		return nil
	}

	payload, ok := evt.Payload.(event.PaymentRequestPayload)
	if !ok {
		return errors.New("invalid payload for PaymentRequested")
	}

	p.Logger.Info("processing payment", map[string]any{
		"invoice-id": payload.InvoiceID,
		"attempt":    payload.Attempt,
	})

	idempotencyKey := generateIdempotencyKey(payload.InvoiceID)

	existing, err := p.Repo.FindByIdempotencyKey(idempotencyKey)
	if err == nil && existing != nil {
		return nil
	}

	paymentID := generatePaymentID()

	pay := &payment.Payment{
		ID:             paymentID,
		InvoiceID:      payload.InvoiceID,
		Attempt:        1,
		Status:         payment.StatusProcessing,
		IdempotencyKey: idempotencyKey,
	}

	if err := p.Repo.Save(pay); err != nil {
		return err
	}

	success := p.Executor.Execute()

	p.Metrics.IncProcessed()

	if success {
		p.Metrics.IncSucceeded()
		p.Logger.Info("payment succeeded", map[string]any{
			"payment-id": pay.ID,
			"invoice-id": payload.InvoiceID,
			"attempt":    payload.Attempt,
		})
		p.Repo.UpdateStatus(pay.ID, payment.StatusSuccess)

		return p.EventBus.Publish(event.Event{
			Type: event.PaymentSucceeded,
			Payload: event.PaymentSucceededPayload{
				InvoiceID: payload.InvoiceID,
				PaymentID: pay.ID,
			},
		})
	}

	p.Metrics.IncFailed()

	p.Logger.Error("payment failed", map[string]any{
		"payment_id": pay.ID,
		"invoice_id": payload.InvoiceID,
		"attempt":    payload.Attempt,
		"retryable":  true,
	})

	p.Repo.UpdateStatus(pay.ID, payment.StatusFailed)

	failPayload := event.PaymentFailedPayload{
		InvoiceID: payload.InvoiceID,
		PaymentID: pay.ID,
		Retryable: true,
		Reason:    "temporary failure",
	}

	p.EventBus.Publish(event.Event{
		Type:    event.PaymentFailed,
		Payload: failPayload,
	})

	p.Retry.Schedule(payload)

	return nil
}
