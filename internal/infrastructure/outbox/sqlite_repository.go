package outbox

import "database/sql"

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db}
}

func (r *SQLiteRepository) Save(evt OutboxEvent) error {
	_, err := r.db.Exec(`
		INSERT INTO outbox_events (id, event_type, payload, published, created_at)
		VALUES (?, ?, ?, ?, ?)
	`,
		evt.ID,
		evt.Type,
		evt.Payload,
		0,
		evt.CreatedAt,
	)
	return err
}

func (r *SQLiteRepository) FindUnpublished(limit int) ([]OutboxEvent, error) {
	rows, err := r.db.Query(`
		SELECT id, event_type, payload, published, created_at
		FROM outbox_events
		WHERE published = 0
		ORDER BY created_at
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []OutboxEvent

	for rows.Next() {
		var evt OutboxEvent
		var published int

		if err := rows.Scan(
			&evt.ID,
			&evt.Type,
			&evt.Payload,
			&published,
			&evt.CreatedAt,
		); err != nil {
			return nil, err
		}

		evt.Published = published == 1
		events = append(events, evt)
	}

	return events, nil
}

func (r *SQLiteRepository) MarkPublished(id string) error {
	_, err := r.db.Exec(`
		UPDATE outbox_events
		SET published = 1
		WHERE id = ?
	`, id)

	return err
}
