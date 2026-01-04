package contracts

import "github.com/rcarvalho-pb/payment_system-go/internal/domain/event"

type EventRecorder interface {
	Record(event.Event) error
}
