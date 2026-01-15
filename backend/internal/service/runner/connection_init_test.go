package runner

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Initialization Tests ==========

func TestHandleInitializeMessage(t *testing.T) {
	t.Run("successful initialization with agent types", func(t *testing.T) {
		// Setup
		logger := newTestLogger()
		cm := NewConnectionManager(logger)
		cm.SetServerVersion("1.0.0")
		cm.SetAgentTypesProvider(&mockAgentTypesProvider{
			agentTypes: []AgentTypeInfo{
				{Slug: "claude-code", Name: "Claude Code", Executable: "claude", LaunchCommand: "claude"},
				{Slug: "aider", Name: "Aider", Executable: "aider", LaunchCommand: "aider"},
			},
		})

		runnerID := int64(123)

		// Add connection to manager
		conn := &RunnerConnection{
			RunnerID: runnerID,
			Conn:     &websocket.Conn{}, // Placeholder, we won't use real WS
			Send:     make(chan []byte, 256),
			LastPing: time.Now(),
		}
		shard := cm.getShard(runnerID)
		shard.mu.Lock()
		shard.connections[runnerID] = conn
		shard.mu.Unlock()

		// Create initialize params
		initParams := InitializeParams{
			ProtocolVersion: CurrentProtocolVersion,
			RunnerInfo: RunnerInfo{
				Version:  "1.0.0",
				NodeID:   "test-node",
				MCPPort:  19000,
				OS:       "linux",
				Arch:     "amd64",
				Hostname: "test-host",
			},
		}
		data, _ := json.Marshal(initParams)

		// Execute
		cm.handleInitializeMessage(runnerID, data)

		// Verify - check that initialize_result was sent via Send channel
		select {
		case msg := <-conn.Send:
			var runnerMsg RunnerMessage
			err := json.Unmarshal(msg, &runnerMsg)
			require.NoError(t, err)
			assert.Equal(t, MsgTypeInitializeResult, runnerMsg.Type)

			var result InitializeResult
			err = json.Unmarshal(runnerMsg.Data, &result)
			require.NoError(t, err)
			assert.Equal(t, CurrentProtocolVersion, result.ProtocolVersion)
			assert.Equal(t, "1.0.0", result.ServerInfo.Version)
			assert.Len(t, result.AgentTypes, 2)
			assert.Contains(t, result.Features, FeatureFilesToCreate)
		case <-time.After(time.Second):
			t.Fatal("expected initialize_result message")
		}
	})

	t.Run("handles empty agent types", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)
		cm.SetServerVersion("1.0.0")
		cm.SetAgentTypesProvider(&mockAgentTypesProvider{agentTypes: nil})

		runnerID := int64(456)
		conn := &RunnerConnection{
			RunnerID: runnerID,
			Conn:     &websocket.Conn{},
			Send:     make(chan []byte, 256),
			LastPing: time.Now(),
		}
		shard := cm.getShard(runnerID)
		shard.mu.Lock()
		shard.connections[runnerID] = conn
		shard.mu.Unlock()

		initParams := InitializeParams{
			ProtocolVersion: CurrentProtocolVersion,
			RunnerInfo:      RunnerInfo{Version: "1.0.0", NodeID: "test"},
		}
		data, _ := json.Marshal(initParams)

		cm.handleInitializeMessage(runnerID, data)

		select {
		case msg := <-conn.Send:
			var runnerMsg RunnerMessage
			json.Unmarshal(msg, &runnerMsg)
			var result InitializeResult
			json.Unmarshal(runnerMsg.Data, &result)
			assert.Empty(t, result.AgentTypes)
		case <-time.After(time.Second):
			t.Fatal("expected initialize_result message")
		}
	})

	t.Run("handles old protocol version with warning", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)
		cm.SetServerVersion("1.0.0")
		cm.SetAgentTypesProvider(&mockAgentTypesProvider{agentTypes: []AgentTypeInfo{}})

		runnerID := int64(789)
		conn := &RunnerConnection{
			RunnerID: runnerID,
			Conn:     &websocket.Conn{},
			Send:     make(chan []byte, 256),
			LastPing: time.Now(),
		}
		shard := cm.getShard(runnerID)
		shard.mu.Lock()
		shard.connections[runnerID] = conn
		shard.mu.Unlock()

		initParams := InitializeParams{
			ProtocolVersion: 1, // Old version
			RunnerInfo:      RunnerInfo{Version: "0.9.0", NodeID: "old-runner"},
		}
		data, _ := json.Marshal(initParams)

		// Should still send response (with logged warning)
		cm.handleInitializeMessage(runnerID, data)

		select {
		case msg := <-conn.Send:
			var runnerMsg RunnerMessage
			json.Unmarshal(msg, &runnerMsg)
			assert.Equal(t, MsgTypeInitializeResult, runnerMsg.Type)
		case <-time.After(time.Second):
			t.Fatal("expected initialize_result message even for old protocol")
		}
	})

	t.Run("handles invalid JSON gracefully", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		// Should not panic
		cm.handleInitializeMessage(123, json.RawMessage(`{invalid json}`))
	})
}

func TestHandleInitializedMessage(t *testing.T) {
	t.Run("updates connection state", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		runnerID := int64(123)
		conn := &RunnerConnection{
			RunnerID: runnerID,
			Conn:     &websocket.Conn{},
			Send:     make(chan []byte, 256),
			LastPing: time.Now(),
		}
		shard := cm.getShard(runnerID)
		shard.mu.Lock()
		shard.connections[runnerID] = conn
		shard.mu.Unlock()

		// Initially not initialized
		assert.False(t, conn.IsInitialized())
		assert.Empty(t, conn.GetAvailableAgents())

		initParams := InitializedParams{
			AvailableAgents: []string{"claude-code", "aider"},
		}
		data, _ := json.Marshal(initParams)

		cm.handleInitializedMessage(runnerID, data)

		// Should now be initialized
		assert.True(t, conn.IsInitialized())
		assert.Equal(t, []string{"claude-code", "aider"}, conn.GetAvailableAgents())
	})

	t.Run("calls onInitialized callback", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		var callbackRunnerID int64
		var callbackAgents []string
		cm.SetInitializedCallback(func(runnerID int64, availableAgents []string) {
			callbackRunnerID = runnerID
			callbackAgents = availableAgents
		})

		runnerID := int64(456)
		conn := &RunnerConnection{
			RunnerID: runnerID,
			Conn:     &websocket.Conn{},
			Send:     make(chan []byte, 256),
		}
		shard := cm.getShard(runnerID)
		shard.mu.Lock()
		shard.connections[runnerID] = conn
		shard.mu.Unlock()

		initParams := InitializedParams{
			AvailableAgents: []string{"gemini-cli"},
		}
		data, _ := json.Marshal(initParams)

		cm.handleInitializedMessage(runnerID, data)

		assert.Equal(t, runnerID, callbackRunnerID)
		assert.Equal(t, []string{"gemini-cli"}, callbackAgents)
	})

	t.Run("handles empty available agents", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		runnerID := int64(789)
		conn := &RunnerConnection{
			RunnerID: runnerID,
			Conn:     &websocket.Conn{},
			Send:     make(chan []byte, 256),
		}
		shard := cm.getShard(runnerID)
		shard.mu.Lock()
		shard.connections[runnerID] = conn
		shard.mu.Unlock()

		initParams := InitializedParams{
			AvailableAgents: []string{},
		}
		data, _ := json.Marshal(initParams)

		cm.handleInitializedMessage(runnerID, data)

		assert.True(t, conn.IsInitialized())
		assert.Empty(t, conn.GetAvailableAgents())
	})
}

func TestHandleInitializeMessage_NewerProtocolVersion(t *testing.T) {
	logger := newTestLogger()
	cm := NewConnectionManager(logger)
	cm.SetServerVersion("1.0.0")
	cm.SetAgentTypesProvider(&mockAgentTypesProvider{agentTypes: []AgentTypeInfo{}})

	runnerID := int64(700)
	conn := &RunnerConnection{
		RunnerID: runnerID,
		Conn:     &websocket.Conn{},
		Send:     make(chan []byte, 256),
		LastPing: time.Now(),
	}
	shard := cm.getShard(runnerID)
	shard.mu.Lock()
	shard.connections[runnerID] = conn
	shard.mu.Unlock()

	// Use a newer protocol version
	initParams := InitializeParams{
		ProtocolVersion: CurrentProtocolVersion + 1, // Newer
		RunnerInfo:      RunnerInfo{Version: "2.0.0", NodeID: "newer-runner"},
	}
	data, _ := json.Marshal(initParams)

	// Should still send response (just log info)
	cm.handleInitializeMessage(runnerID, data)

	select {
	case msg := <-conn.Send:
		var runnerMsg RunnerMessage
		json.Unmarshal(msg, &runnerMsg)
		assert.Equal(t, MsgTypeInitializeResult, runnerMsg.Type)
	case <-time.After(time.Second):
		t.Fatal("expected initialize_result message for newer protocol")
	}
}

func TestHandleInitializeMessage_NoAgentTypesProvider(t *testing.T) {
	logger := newTestLogger()
	cm := NewConnectionManager(logger)
	cm.SetServerVersion("1.0.0")
	// No agent types provider set

	runnerID := int64(701)
	conn := &RunnerConnection{
		RunnerID: runnerID,
		Conn:     &websocket.Conn{},
		Send:     make(chan []byte, 256),
		LastPing: time.Now(),
	}
	shard := cm.getShard(runnerID)
	shard.mu.Lock()
	shard.connections[runnerID] = conn
	shard.mu.Unlock()

	initParams := InitializeParams{
		ProtocolVersion: CurrentProtocolVersion,
		RunnerInfo:      RunnerInfo{Version: "1.0.0", NodeID: "test"},
	}
	data, _ := json.Marshal(initParams)

	cm.handleInitializeMessage(runnerID, data)

	select {
	case msg := <-conn.Send:
		var runnerMsg RunnerMessage
		json.Unmarshal(msg, &runnerMsg)
		var result InitializeResult
		json.Unmarshal(runnerMsg.Data, &result)
		// Should have empty agent types
		assert.Empty(t, result.AgentTypes)
	case <-time.After(time.Second):
		t.Fatal("expected initialize_result message")
	}
}

func TestHandleInitializedMessage_InvalidJSON(t *testing.T) {
	logger := newTestLogger()
	cm := NewConnectionManager(logger)

	callbackCalled := false
	cm.SetInitializedCallback(func(runnerID int64, availableAgents []string) {
		callbackCalled = true
	})

	cm.handleInitializedMessage(800, json.RawMessage(`{invalid}`))
	assert.False(t, callbackCalled)
}

func TestHandleInitializedMessage_NoConnection(t *testing.T) {
	logger := newTestLogger()
	cm := NewConnectionManager(logger)

	callbackCalled := false
	cm.SetInitializedCallback(func(runnerID int64, availableAgents []string) {
		callbackCalled = true
	})

	// No connection exists for this runner ID
	initParams := InitializedParams{
		AvailableAgents: []string{"claude-code"},
	}
	data, _ := json.Marshal(initParams)

	// Should still call callback but not panic when updating connection
	cm.handleInitializedMessage(999, data)

	assert.True(t, callbackCalled)
}

// ========== Initialization Failure Tests ==========

func TestFailInitialization(t *testing.T) {
	t.Run("calls callback and removes connection", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		var callbackRunnerID int64
		var callbackReason string
		cm.SetInitFailedCallback(func(runnerID int64, reason string) {
			callbackRunnerID = runnerID
			callbackReason = reason
		})

		runnerID := int64(123)

		// Add a connection
		conn := newTestWebSocketConn(t)
		cm.AddConnection(runnerID, conn)
		require.True(t, cm.IsConnected(runnerID))

		// Trigger failure
		cm.failInitialization(runnerID, "test failure reason")

		// Verify callback was called
		assert.Equal(t, runnerID, callbackRunnerID)
		assert.Equal(t, "test failure reason", callbackReason)

		// Verify connection was removed
		assert.False(t, cm.IsConnected(runnerID))
	})

	t.Run("handles no callback set", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		runnerID := int64(456)

		// Add a connection
		conn := newTestWebSocketConn(t)
		cm.AddConnection(runnerID, conn)

		// Should not panic when callback is not set
		cm.failInitialization(runnerID, "test reason")

		// Connection should still be removed
		assert.False(t, cm.IsConnected(runnerID))
	})
}

func TestSetInitFailedCallback(t *testing.T) {
	logger := newTestLogger()
	cm := NewConnectionManager(logger)

	called := false
	cm.SetInitFailedCallback(func(runnerID int64, reason string) {
		called = true
	})

	// Verify callback is set (we can trigger it via failInitialization)
	cm.failInitialization(1, "test")
	assert.True(t, called)
}

func TestHandleInitializeMessage_InvalidJSON_CallsFailCallback(t *testing.T) {
	logger := newTestLogger()
	cm := NewConnectionManager(logger)

	var failedRunnerID int64
	var failReason string
	cm.SetInitFailedCallback(func(runnerID int64, reason string) {
		failedRunnerID = runnerID
		failReason = reason
	})

	runnerID := int64(789)
	conn := newTestWebSocketConn(t)
	cm.AddConnection(runnerID, conn)

	// Send invalid JSON
	cm.handleInitializeMessage(runnerID, json.RawMessage(`{invalid json}`))

	// Verify failure callback was called
	assert.Equal(t, runnerID, failedRunnerID)
	assert.Contains(t, failReason, "invalid initialize params")

	// Verify connection was removed
	assert.False(t, cm.IsConnected(runnerID))
}

func TestInitTimeout_CallsFailCallback(t *testing.T) {
	logger := newTestLogger()
	cm := NewConnectionManager(logger)
	cm.SetInitTimeout(50 * time.Millisecond)

	var failedRunnerIDs []int64
	var mu sync.Mutex
	cm.SetInitFailedCallback(func(runnerID int64, reason string) {
		mu.Lock()
		failedRunnerIDs = append(failedRunnerIDs, runnerID)
		mu.Unlock()
		assert.Equal(t, "initialization timeout", reason)
	})

	// Add uninitialized connection
	conn := newTestWebSocketConn(t)
	rc := cm.AddConnection(1, conn)
	require.False(t, rc.IsInitialized())

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// Manually trigger check
	cm.checkInitTimeouts()

	// Verify callback was called
	mu.Lock()
	assert.Contains(t, failedRunnerIDs, int64(1))
	mu.Unlock()

	// Verify connection was removed
	assert.False(t, cm.IsConnected(1))
}

func TestInitTimeout_DoesNotAffectInitializedConnections(t *testing.T) {
	logger := newTestLogger()
	cm := NewConnectionManager(logger)
	cm.SetInitTimeout(50 * time.Millisecond)

	failCallbackCalled := false
	cm.SetInitFailedCallback(func(runnerID int64, reason string) {
		failCallbackCalled = true
	})

	// Add connection and mark as initialized
	conn := newTestWebSocketConn(t)
	rc := cm.AddConnection(1, conn)
	rc.SetInitialized(true, []string{"claude-code"})

	// Wait for timeout period
	time.Sleep(100 * time.Millisecond)

	// Trigger check
	cm.checkInitTimeouts()

	// Callback should NOT have been called
	assert.False(t, failCallbackCalled)

	// Connection should still exist
	assert.True(t, cm.IsConnected(1))
}

func TestProtocolVersionConstants(t *testing.T) {
	t.Run("version constants are valid", func(t *testing.T) {
		assert.GreaterOrEqual(t, CurrentProtocolVersion, MinSupportedProtocolVersion)
		assert.Equal(t, 2, CurrentProtocolVersion)
		assert.Equal(t, 2, MinSupportedProtocolVersion)
	})

	t.Run("supported features are defined", func(t *testing.T) {
		features := SupportedFeatures()
		assert.Contains(t, features, FeatureFilesToCreate)
		assert.Contains(t, features, FeatureWorkDirConfig)
		assert.Contains(t, features, FeatureInitialPrompt)
	})
}
