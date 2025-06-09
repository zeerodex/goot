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
	google_id TEXT,
	todoist_id TEXT,
	title TEXT,
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

	gtasksSnapshotsStmt := `
	CREATE TABLE IF NOT EXISTS gtasks_snapshots (
	id INTEGER PRIMARY KEY,
	api_id TEXT,
	title TEXT,
	description TEXT,
	due TEXT,
	last_modified TEXT,
	completed BOOl DEFAULT 0)`

	_, err = db.Exec(gtasksSnapshotsStmt)
	if err != nil {
		return nil, fmt.Errorf("failed to create gtasks_snapshots table: %w", err)
	}

	todoistSnapshotsStmt := `
	CREATE TABLE IF NOT EXISTS todoist_snapshots (
	id INTEGER PRIMARY KEY,
	api_ID TEXT,
	title TEXT,
	description TEXT,
	due TEXT,
	last_modified TEXT,
	completed BOOl DEFAULT 0)`

	_, err = db.Exec(todoistSnapshotsStmt)
	if err != nil {
		return nil, fmt.Errorf("failed to create todoist_snapshots table: %w", err)
	}

	return db, nil
}
