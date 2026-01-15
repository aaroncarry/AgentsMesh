package runner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunnerConnectionSendMessage(t *testing.T) {
	t.Run("sends message successfully", func(t *testing.T) {
		conn := newTestWebSocketConn(t)
		rc := &RunnerConnection{
			RunnerID: 1,
			Conn:     conn,
			Send:     make(chan []byte, 256),
			LastPing: time.Now(),
		}

		msg := &RunnerMessage{
			Type:      MsgTypeHeartbeat,
			Timestamp: time.Now().UnixMilli(),
		}

		err := rc.SendMessage(msg)
		require.NoError(t, err)

		// Should receive message on send channel
		select {
		case data := <-rc.Send:
			assert.NotEmpty(t, data)
		case <-time.After(time.Second):
			t.Fatal("expected message on send channel")
		}
	})

	t.Run("returns error when connection is nil", func(t *testing.T) {
		rc := &RunnerConnection{
			RunnerID: 1,
			Conn:     nil, // nil connection
			Send:     make(chan []byte, 256),
			LastPing: time.Now(),
		}

		msg := &RunnerMessage{
			Type: MsgTypeHeartbeat,
		}

		err := rc.SendMessage(msg)
		assert.Equal(t, ErrConnectionClosed, err)
	})

	t.Run("returns error when send buffer is full", func(t *testing.T) {
		conn := newTestWebSocketConn(t)
		rc := &RunnerConnection{
			RunnerID: 1,
			Conn:     conn,
			Send:     make(chan []byte, 1), // Very small buffer
			LastPing: time.Now(),
		}

		msg := &RunnerMessage{
			Type: MsgTypeHeartbeat,
		}

		// Fill the buffer
		rc.Send <- []byte("blocking message")

		// Now try to send another - should fail
		err := rc.SendMessage(msg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "buffer full")
	})
}

func TestRunnerConnectionWritePump(t *testing.T) {
	t.Run("writes messages from send channel", func(t *testing.T) {
		conn := newTestWebSocketConn(t)
		rc := &RunnerConnection{
			RunnerID:     1,
			Conn:         conn,
			Send:         make(chan []byte, 256),
			LastPing:     time.Now(),
			PingInterval: 100 * time.Millisecond,
		}

		// Start write pump in goroutine
		done := make(chan struct{})
		go func() {
			rc.WritePump()
			close(done)
		}()

		// Send a message
		rc.Send <- []byte(`{"type":"heartbeat"}`)

		// Give it time to process
		time.Sleep(50 * time.Millisecond)

		// Use Close() instead of close(rc.Send) - WritePump.defer calls Close() which handles channel
		rc.mu.Lock()
		rc.Conn = nil // Set conn to nil to trigger exit
		rc.mu.Unlock()

		// Wait for write pump to finish
		select {
		case <-done:
			// Success
		case <-time.After(time.Second):
			t.Fatal("write pump did not finish")
		}
	})

	t.Run("handles nil connection", func(t *testing.T) {
		rc := &RunnerConnection{
			RunnerID:     1,
			Conn:         nil, // nil connection
			Send:         make(chan []byte, 256),
			LastPing:     time.Now(),
			PingInterval: 50 * time.Millisecond,
		}

		done := make(chan struct{})
		go func() {
			rc.WritePump()
			close(done)
		}()

		// Send a message - should exit because conn is nil
		rc.Send <- []byte(`{"type":"test"}`)

		select {
		case <-done:
			// Success - pump exited
		case <-time.After(time.Second):
			t.Fatal("write pump did not exit when connection is nil")
		}
	})

	t.Run("uses default ping interval when not set", func(t *testing.T) {
		conn := newTestWebSocketConn(t)
		rc := &RunnerConnection{
			RunnerID:     1,
			Conn:         conn,
			Send:         make(chan []byte, 256),
			LastPing:     time.Now(),
			PingInterval: 0, // Zero - should use default
		}

		done := make(chan struct{})
		go func() {
			rc.WritePump()
			close(done)
		}()

		// Close immediately
		rc.Close()

		select {
		case <-done:
			// Success
		case <-time.After(time.Second):
			t.Fatal("write pump did not finish")
		}
	})
}

func TestRunnerConnectionClose(t *testing.T) {
	t.Run("closes connection and channel", func(t *testing.T) {
		conn := newTestWebSocketConn(t)
		rc := &RunnerConnection{
			RunnerID: 1,
			Conn:     conn,
			Send:     make(chan []byte, 10),
			LastPing: time.Now(),
		}

		rc.Close()

		// Verify connection is nil
		rc.mu.Lock()
		assert.Nil(t, rc.Conn)
		rc.mu.Unlock()

		// Verify channel is closed
		_, ok := <-rc.Send
		assert.False(t, ok, "send channel should be closed")
	})

	t.Run("is idempotent", func(t *testing.T) {
		conn := newTestWebSocketConn(t)
		rc := &RunnerConnection{
			RunnerID: 1,
			Conn:     conn,
			Send:     make(chan []byte, 10),
			LastPing: time.Now(),
		}

		// Multiple closes should not panic
		rc.Close()
		rc.Close()
		rc.Close()
	})
}

func TestRunnerConnectionState(t *testing.T) {
	t.Run("IsInitialized returns false by default", func(t *testing.T) {
		rc := &RunnerConnection{}
		assert.False(t, rc.IsInitialized())
	})

	t.Run("SetInitialized updates state", func(t *testing.T) {
		rc := &RunnerConnection{}
		rc.SetInitialized(true, []string{"claude-code", "aider"})

		assert.True(t, rc.IsInitialized())
		assert.Equal(t, []string{"claude-code", "aider"}, rc.GetAvailableAgents())
	})

	t.Run("GetAvailableAgents returns empty slice by default", func(t *testing.T) {
		rc := &RunnerConnection{}
		assert.Empty(t, rc.GetAvailableAgents())
	})

	t.Run("state access is thread-safe", func(t *testing.T) {
		rc := &RunnerConnection{}

		done := make(chan struct{})

		// Writer goroutine
		go func() {
			for i := 0; i < 100; i++ {
				rc.SetInitialized(true, []string{"agent"})
			}
			done <- struct{}{}
		}()

		// Reader goroutine
		go func() {
			for i := 0; i < 100; i++ {
				_ = rc.IsInitialized()
				_ = rc.GetAvailableAgents()
			}
			done <- struct{}{}
		}()

		// Wait for both
		<-done
		<-done
	})
}
