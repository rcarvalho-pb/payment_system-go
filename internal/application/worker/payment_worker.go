package worker

import "github.com/rcarvalho-pb/payment_system-go/internal/domain/event"

type PaymentWorker struct {
	Handler PaymentHandler
}

type PaymentHandler interface {
	Handle(event.Event) error
}

type PaymentExecutor interface {
	Execute() bool
}
