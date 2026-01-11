package runner

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// Create tables manually for SQLite compatibility
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS runner_registration_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			organization_id INTEGER NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			description TEXT,
			created_by_id INTEGER NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1,
			max_uses INTEGER,
			used_count INTEGER NOT NULL DEFAULT 0,
			expires_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		t.Fatalf("failed to create registration_tokens table: %v", err)
	}

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS runners (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			organization_id INTEGER NOT NULL,
			node_id TEXT NOT NULL,
			description TEXT,
			auth_token_hash TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'offline',
			last_heartbeat DATETIME,
			current_pods INTEGER NOT NULL DEFAULT 0,
			max_concurrent_pods INTEGER NOT NULL DEFAULT 5,
			runner_version TEXT,
			is_enabled INTEGER NOT NULL DEFAULT 1,
			host_info TEXT,
			capabilities TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		t.Fatalf("failed to create runners table: %v", err)
	}

	return db
}
