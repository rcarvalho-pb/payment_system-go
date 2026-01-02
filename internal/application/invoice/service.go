package invoice

import (
	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_system-go/internal/domain/invoice"
)

type Service struct {
	Repo     invoice.Repository
	EventBus EventPublisher
}

type EventPublisher interface {
	Publish(event.Event) error
}
