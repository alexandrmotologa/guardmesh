package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	// Optimize SQLite for high concurrency reverse proxy logging
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("exec pragma %s: %w", pragma, err)
		}
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate sqlite schemas: %w", err)
	}

	return s, nil
}

func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS audit_events (
		id TEXT PRIMARY KEY,
		timestamp DATETIME NOT NULL,
		app_id TEXT NOT NULL,
		client_ip TEXT NOT NULL,
		provider TEXT NOT NULL,
		model TEXT NOT NULL,
		status_code INTEGER NOT NULL,
		latency_ms REAL NOT NULL,
		prompt_tokens INTEGER NOT NULL,
		completion_tokens INTEGER NOT NULL,
		risk_tier TEXT NOT NULL,
		risk_score REAL NOT NULL,
		risk_reasons TEXT NOT NULL,
		is_blocked INTEGER NOT NULL,
		blocked_reason TEXT,
		pii_detected_count INTEGER NOT NULL,
		pii_categories TEXT NOT NULL,
		prompt_sha256 TEXT NOT NULL,
		sample_snippet TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_events(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_audit_app_id ON audit_events(app_id);
	CREATE INDEX IF NOT EXISTS idx_audit_risk_tier ON audit_events(risk_tier);
	CREATE INDEX IF NOT EXISTS idx_audit_is_blocked ON audit_events(is_blocked);

	CREATE TABLE IF NOT EXISTS applications (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		first_seen DATETIME NOT NULL,
		last_seen DATETIME NOT NULL,
		highest_risk_tier TEXT NOT NULL,
		total_requests INTEGER NOT NULL DEFAULT 0,
		total_pii_masked INTEGER NOT NULL DEFAULT 0,
		compliance_status TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS dossier_records (
		id TEXT PRIMARY KEY,
		app_id TEXT NOT NULL,
		generated_at DATETIME NOT NULL,
		risk_tier TEXT NOT NULL,
		file_path TEXT NOT NULL,
		summary TEXT NOT NULL
	);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) Ping() error {
	return s.db.Ping()
}
