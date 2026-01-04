package outbox_test

import (
	"errors"
	"testing"
	"time"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_system-go/internal/infrastructure/outbox"
)

type fakeBus struct {
	published []event.Event
	fail      bool
}

func (f *fakeBus) Publish(evt event.Event) error {
	if f.fail {
		return errors.New("bus down")
	}
	f.published = append(f.published, evt)
	return nil
}

func TestDispatcher_ShouldPublishAndMarkEvent(t *testing.T) {
	db := setupTestDB(t)
	repo := outbox.NewSQLiteRepository(db)

	bus := &fakeBus{}

	dispatcher := &outbox.Dispatcher{
		Repo:         repo,
		EventBus:     bus,
		PollInterval: time.Millisecond,
		BatchSize:    10,
	}

	payload := []byte(`{"invoice_id":"inv-1","payment_id":"pay-1"}`)

	err := repo.Save(outbox.OutboxEvent{
		ID:        "evt-1",
		Type:      event.PaymentSucceeded,
		Payload:   payload,
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	dispatcher.DispatchOnce()

	if len(bus.published) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(bus.published))
	}

	events, _ := repo.FindUnpublished(10)
	if len(events) != 0 {
		t.Fatalf("expected no unpublished events")
	}
}
