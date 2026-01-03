package invoice

import (
	"errors"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
	domainInvoice "github.com/rcarvalho-pb/payment_system-go/internal/domain/invoice"
)

type PaymentEventHandler struct {
	Repo domainInvoice.Repository
}

func (h *PaymentEventHandler) Handle(evt event.Event) error {
	switch evt.Type {
	case event.PaymentSucceeded:
		payload, ok := evt.Payload.(event.PaymentSucceededPayload)
		if !ok {
			return errors.New("invalid payload for PaymentSucceeded")
		}
		return h.Repo.UpdateStatus(payload.InvoiceID, domainInvoice.StatusPaid)

	case event.PaymentFailed:
		payload, ok := evt.Payload.(event.PaymentFailedPayload)
		if !ok {
			return errors.New("invalid payload for PaymentFailed")
		}
		if !payload.Retryable {
			return h.Repo.UpdateStatus(payload.InvoiceID, domainInvoice.StatusFailed)
		}
		return nil
	}
	return nil
}
