package outbox

import (
	"time"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
)

type OutboxEvent struct {
	ID        string
	Type      event.Type
	Payload   []byte
	Published bool
	CreatedAt time.Time
}

type Repository interface {
	Save(OutboxEvent) error
	FindUnpublished(int) ([]OutboxEvent, error)
	MarkPublished(string) error
}
