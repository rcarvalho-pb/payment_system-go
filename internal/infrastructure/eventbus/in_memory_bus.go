package eventbus

import (
	"sync"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
)

type HandlerFunc func(event.Event) error

type InMemoryBus struct {
	mu       sync.RWMutex
	handlers map[event.Type][]HandlerFunc
}

func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{
		handlers: make(map[event.Type][]HandlerFunc),
	}
}

func (b *InMemoryBus) Subscribe(eventType event.Type, handler HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *InMemoryBus) Publish(evt event.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	handlers := b.handlers[evt.Type]

	for _, handler := range handlers {
		if err := handler(evt); err != nil {
			return err
		}
	}

	return nil
}
