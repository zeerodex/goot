package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	stmt := `
	CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY,
	google_id TEXT,
	title TEXT,
	description TEXT,
	due TEXT,
	last_modified TEXT,
	completed BOOl DEFAULT 0,
	deleted BOOLEAN DEFAULT 0,
	notified BOOLEAN DEFAULT 0)`

	_, err = db.Exec(stmt)
	if err != nil {
		return nil, err
	}

	return db, nil
}
