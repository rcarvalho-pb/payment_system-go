package sqlite

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// pragmas := []string{
	// 	"PRAGMA journal_mode = WAL;",
	// 	"PRAGMA foreign_keys = ON;",
	// 	"PRAGMA synchronous = NORMAL;",
	// }

	// for _, p := range pragmas {
	// 	if _, err := db.Exec(p); err != nil {
	// 		return nil, err
	// 	}
	// }

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
