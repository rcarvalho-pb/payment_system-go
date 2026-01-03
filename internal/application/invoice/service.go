package invoice

import (
	"errors"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
	domainInvoice "github.com/rcarvalho-pb/payment_system-go/internal/domain/invoice"
)

var (
	ErrInvoiceNotFound     = errors.New("invoice not found")
	ErrInvalidInvoiceState = errors.New("invalid invoice state")
)

type Service struct {
	Repo     domainInvoice.Repository
	EventBus EventPublisher
}

type EventPublisher interface {
	Publish(event.Event) error
}

func (s *Service) CreateInvoice(id string, amount int64) (*domainInvoice.Invoice, error) {
	inv := &domainInvoice.Invoice{
		ID:     id,
		Amount: amount,
		Status: domainInvoice.StatusPending,
	}

	if err := s.Repo.Save(inv); err != nil {
		return nil, err
	}

	return inv, nil
}

func (s *Service) RequestPayment(invoiceID string) error {
	inv, err := s.Repo.FindByID(invoiceID)
	if err != nil {
		return err
	}

	if inv.Status != domainInvoice.StatusPending {
		return ErrInvalidInvoiceState
	}

	if err := s.Repo.UpdateStatus(inv.ID, domainInvoice.StatusProcessing); err != nil {
		return err
	}

	evt := event.Event{
		Type: event.PaymentRequested,
		Payload: event.PaymentRequestPayload{
			InvoiceID: inv.ID,
			Amount:    inv.Amount,
			Attempt:   1,
		},
	}

	return s.EventBus.Publish(evt)
}
