package payment

import (
	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_system-go/internal/domain/payment"
)

type Service struct {
	Repo     payment.Repository
	EventBus EventPublisher
}

type EventPublisher interface {
	Publish(event.Event) error
}
