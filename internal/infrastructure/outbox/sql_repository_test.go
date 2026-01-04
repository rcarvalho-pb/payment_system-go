package outbox_test

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_system-go/internal/infrastructure/outbox"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	schema := `
	CREATE TABLE outbox_events (
		id TEXT PRIMARY KEY,
		event_type TEXT NOT NULL,
		payload BLOB NOT NULL,
		published INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatal(err)
	}

	return db
}

func TestOutbox_ShouldPersistEvent_BeforePublish(t *testing.T) {
	db := setupTestDB(t)
	repo := outbox.NewSQLiteRepository(db)

	evt := outbox.OutboxEvent{
		ID:        "evt-1",
		Type:      event.PaymentSucceeded,
		Payload:   []byte(`{"invoice_id":"inv-1"}`),
		CreatedAt: time.Now(),
	}

	err := repo.Save(evt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events, err := repo.FindUnpublished(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Published {
		t.Fatalf("expected event to be unpublished")
	}
}
