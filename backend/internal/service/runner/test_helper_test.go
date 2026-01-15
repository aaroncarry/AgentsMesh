package runner

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// testWebSocketServer creates a test WebSocket server and returns the client connection
// The server is closed when the test completes
func newTestWebSocketConn(t *testing.T) *websocket.Conn {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	// Create a test server that accepts WebSocket connections
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade error: %v", err)
			return
		}
		// Keep connection alive by reading messages until closed
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))

	t.Cleanup(func() {
		server.Close()
	})

	// Connect to the test server
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect to test WebSocket server: %v", err)
	}

	t.Cleanup(func() {
		conn.Close()
	})

	return conn
}

// newMockWebsocketConn creates a mock WebSocket connection for testing
// DEPRECATED: Use newTestWebSocketConn for tests that need a real connection
func newMockWebsocketConn() *websocket.Conn {
	// This returns nil - tests using this should be updated to use newTestWebSocketConn
	return nil
}

// newTestLogger creates a test logger that only logs errors
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// Create runner_registration_tokens table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS runner_registration_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			organization_id INTEGER NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			description TEXT,
			created_by_id INTEGER NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			max_uses INTEGER,
			used_count INTEGER NOT NULL DEFAULT 0,
			expires_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		t.Fatalf("failed to create runner_registration_tokens table: %v", err)
	}

	// Create runners table
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
			is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
			host_info TEXT,
			available_agents TEXT DEFAULT '[]',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		t.Fatalf("failed to create runners table: %v", err)
	}

	// Create indexes
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_runners_organization_id ON runners(organization_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_runners_status ON runners(status)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_runner_registration_tokens_organization_id ON runner_registration_tokens(organization_id)`)

	return db
}
