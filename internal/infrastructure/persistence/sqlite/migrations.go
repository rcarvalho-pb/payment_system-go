package sqlite

import "database/sql"

func RunMigrations(db *sql.DB) error {
	stmts := []string{

		`CREATE TABLE IF NOT EXISTS invoices (
			id TEXT PRIMARY KEY,
			amount INTEGER NOT NULL,
			status TEXT NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS payments (
			id TEXT PRIMARY KEY,
			invoice_id TEXT NOT NULL,
			attempt INTEGER NOT NULL,
			status TEXT NOT NULL,
			idempotency_key TEXT NOT NULL UNIQUE
		);`,

		`CREATE TABLE IF NOT EXISTS outbox_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event_type TEXT NOT NULL,
			payload TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			published_at DATETIME
		);`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
