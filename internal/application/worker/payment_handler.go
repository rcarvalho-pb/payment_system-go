package worker

import (
	"errors"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_system-go/internal/domain/payment"
	"honnef.co/go/tools/analysis/facts/generated"
)

type PaymentProcessor struct {
	Repo     payment.Repository
	EventBus EventPublisher
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

	success := simulatePayment()

	if success {
		p.Repo.UpdateStatus(pay.ID, payment.StatusSuccess)

		return p.EventBus.Publish(event.Event{
			Type: event.PaymentSucceeded,
			Payload: event.PaymentSucceededPayload{
				InvoiceID: payload.InvoiceID,
				PaymentID: pay.ID,
			},
		})
	}

	p.Repo.UpdateStatus(pay.ID, payment.StatusFailed)

	return p.EventBus.Publish(event.Event{
		Type: event.PaymentFailed,
		Payload: event.PaymentFailedPayload{
			InvoiceID: payload.InvoiceID,
			PaymentID: pay.ID,
			Retryable: true,
			Reason:    "temporary failure",
		},
	})
}
