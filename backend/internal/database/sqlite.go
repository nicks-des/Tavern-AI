package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "tavern.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	// Enable WAL mode and foreign keys
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	} {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("pragma %s: %w", pragma, err)
		}
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &DB{db}, nil
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS characters (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			avatar TEXT DEFAULT '',
			portrait TEXT DEFAULT '',
			catchphrase TEXT DEFAULT '',
			description TEXT NOT NULL,
			personality TEXT DEFAULT '',
			scenario TEXT DEFAULT '',
			first_message TEXT DEFAULT '',
			example_dialogue TEXT DEFAULT '',
			tags TEXT DEFAULT '[]',
			scope TEXT NOT NULL DEFAULT 'global',
			room_id TEXT,
			created_at DATETIME NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE TABLE IF NOT EXISTS rooms (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			world_rules TEXT DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT (datetime('now')),
			updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE TABLE IF NOT EXISTS room_members (
			room_id TEXT NOT NULL,
			character_id TEXT NOT NULL,
			overrides TEXT DEFAULT '{}',
			PRIMARY KEY (room_id, character_id),
			FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
			FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			character_id TEXT NOT NULL,
			room_id TEXT,
			title TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT (datetime('now')),
			updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL CHECK (role IN ('user','assistant','system')),
			content TEXT NOT NULL DEFAULT '',
			timestamp DATETIME NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
		)`,

		`CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_character ON sessions(character_id)`,
		`CREATE INDEX IF NOT EXISTS idx_characters_scope ON characters(scope)`,
	}

	for i, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration %d: %w", i, err)
		}
	}

	return nil
}
