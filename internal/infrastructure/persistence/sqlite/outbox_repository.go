package sqlite

import (
	"database/sql"
	"encoding/json"

	"github.com/rcarvalho-pb/payment_system-go/internal/domain/event"
)

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) Save(evt event.Event) error {
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

func (r *OutboxRepository) FindUnpublished(limit int) ([]event.Event, []int64, error) {
	rows, err := r.db.Query(
		`SELECT id, event_type, payload
		 FROM outbox_events
		 WHERE published = 0
		 ORDER BY id
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var events []event.Event
	var ids []int64

	for rows.Next() {
		var (
			id      int64
			typ     string
			payload []byte
		)

		if err := rows.Scan(&id, &typ, &payload); err != nil {
			return nil, nil, err
		}

		evt := event.Event{
			Type:    event.Type(typ),
			Payload: payload, // deserializado no dispatcher
		}

		events = append(events, evt)
		ids = append(ids, id)
	}

	return events, ids, nil
}

func (r *OutboxRepository) MarkPublished(ids []int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, id := range ids {
		if _, err := tx.Exec(
			`UPDATE outbox_events SET published = 1 WHERE id = ?`,
			id,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}
