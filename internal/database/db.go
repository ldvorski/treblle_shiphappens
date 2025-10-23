package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) RunMigrations() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS api_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			method TEXT NOT NULL,
			path TEXT NOT NULL,
			response_status INTEGER NOT NULL,
			response_time_ms INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_created_at ON api_requests(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_response_time ON api_requests(response_time_ms DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_method ON api_requests(method)`,
		`CREATE INDEX IF NOT EXISTS idx_response_status ON api_requests(response_status)`,
		`CREATE TABLE IF NOT EXISTS problems (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			request_id INTEGER NOT NULL,
			problem_type TEXT NOT NULL,
			description TEXT NOT NULL,
			threshold_ms INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (request_id) REFERENCES api_requests(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_problem_created_at ON problems(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_problem_request_id ON problems(request_id)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
