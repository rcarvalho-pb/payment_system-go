package sqlite

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_system-go/internal/infrastructure/outbox"
)

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

// Save(OutboxEvent) error
// FindUnpublished(int) ([]OutboxEvent, error)
// MarkPublished(string) error

func (r *OutboxRepository) Save(evt outbox.OutboxEvent) error {
	data, err := json.Marshal(evt.Payload)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO outbox_events (event_type, payload)
		 VALUES (?, ?)`,
		string(evt.Type),
		data,
	)
	return err
}

func (r *OutboxRepository) FindUnpublished(limit int) ([]outbox.OutboxEvent, error) {
	rows, err := r.db.Query(
		`SELECT id, event_type, payload
		 FROM outbox_events
		 WHERE published = 0
		 ORDER BY id
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []outbox.OutboxEvent
	var ids []string

	for rows.Next() {
		var (
			id      string
			typ     string
			payload []byte
		)

		if err := rows.Scan(&id, &typ, &payload); err != nil {
			return nil, err
		}

		evt := outbox.OutboxEvent{
			ID:        id,
			Type:      event.Type(typ),
			Payload:   payload,
			Published: false,
			CreatedAt: time.Now(),
		}

		events = append(events, evt)
		ids = append(ids, id)
	}

	return events, nil
}

func (r *OutboxRepository) MarkPublished(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`UPDATE outbox_events SET published = 1 WHERE id = ?`,
		id,
	); err != nil {
		return err
	}

	return tx.Commit()
}
