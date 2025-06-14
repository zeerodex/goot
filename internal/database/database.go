package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		return nil, err
	}

	stmt := `
	CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY,
	api_ids TEXT,
	title TEXT NOT NULL,
	description TEXT,
	due TEXT,
	last_modified TEXT,
	completed BOOl DEFAULT 0,
	deleted BOOLEAN DEFAULT 0,
	notified BOOLEAN DEFAULT 0)`

	_, err = db.Exec(stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to create tasks table: %w", err)
	}

	snapshotsStmt := `
	CREATE TABLE IF NOT EXISTS snapshots (
    id INTEGER PRIMARY KEY,
    api TEXT NOT NULL, 
    timestamp TEXT NOT NULL,
    api_ids TEXT)`

	_, err = db.Exec(snapshotsStmt)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshots table: %w", err)
	}

	return db, nil
}
