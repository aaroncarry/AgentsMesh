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

// mockAgentTypesProvider implements AgentTypesProvider for testing
type mockAgentTypesProvider struct {
	agentTypes []AgentTypeInfo
}

func (m *mockAgentTypesProvider) GetAgentTypesForRunner() []AgentTypeInfo {
	return m.agentTypes
}

// mockWebSocketConn implements a minimal WebSocket connection for testing
type mockWebSocketConn struct {
	mu           sync.Mutex
	sentMessages [][]byte
	closed       bool
}

func (m *mockWebSocketConn) WriteMessage(messageType int, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = append(m.sentMessages, data)
	return nil
}

func (m *mockWebSocketConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockWebSocketConn) getSentMessages() [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([][]byte{}, m.sentMessages...)
}

// ========== Message Handling Tests ==========

func TestHandleMessage(t *testing.T) {
	t.Run("handles text message type", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		var receivedData *HeartbeatData
		cm.SetHeartbeatCallback(func(runnerID int64, data *HeartbeatData) {
			receivedData = data
		})

		// Add connection
		runnerID := int64(100)
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

		hbData := HeartbeatData{
			Pods: []HeartbeatPod{{PodKey: "pod-1", Status: "running"}},
		}
		hbJSON, _ := json.Marshal(hbData)

		msg := RunnerMessage{
			Type: MsgTypeHeartbeat,
			Data: hbJSON,
		}
		msgData, _ := json.Marshal(msg)

		cm.HandleMessage(runnerID, websocket.TextMessage, msgData)

		require.NotNil(t, receivedData)
		assert.Len(t, receivedData.Pods, 1)
		assert.Equal(t, "pod-1", receivedData.Pods[0].PodKey)
	})

	t.Run("handles binary message type", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		var receivedData *HeartbeatData
		cm.SetHeartbeatCallback(func(runnerID int64, data *HeartbeatData) {
			receivedData = data
		})

		// Add connection
		runnerID := int64(101)
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

		hbData := HeartbeatData{Pods: []HeartbeatPod{}}
		hbJSON, _ := json.Marshal(hbData)

		msg := RunnerMessage{
			Type: MsgTypeHeartbeat,
			Data: hbJSON,
		}
		msgData, _ := json.Marshal(msg)

		cm.HandleMessage(runnerID, websocket.BinaryMessage, msgData)

		require.NotNil(t, receivedData)
	})

	t.Run("ignores ping message type", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		called := false
		cm.SetHeartbeatCallback(func(runnerID int64, data *HeartbeatData) {
			called = true
		})

		cm.HandleMessage(100, websocket.PingMessage, []byte{})
		assert.False(t, called)
	})

	t.Run("ignores pong message type", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		called := false
		cm.SetHeartbeatCallback(func(runnerID int64, data *HeartbeatData) {
			called = true
		})

		cm.HandleMessage(100, websocket.PongMessage, []byte{})
		assert.False(t, called)
	})

	t.Run("ignores close message type", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		called := false
		cm.SetHeartbeatCallback(func(runnerID int64, data *HeartbeatData) {
			called = true
		})

		cm.HandleMessage(100, websocket.CloseMessage, []byte{})
		assert.False(t, called)
	})

	t.Run("handles invalid JSON gracefully", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		// Should not panic
		cm.HandleMessage(100, websocket.TextMessage, []byte(`{not valid json}`))
	})

	t.Run("handles unknown message type", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		msg := RunnerMessage{
			Type: "unknown_type",
			Data: json.RawMessage(`{}`),
		}
		msgData, _ := json.Marshal(msg)

		// Should not panic and should just log
		cm.HandleMessage(100, websocket.TextMessage, msgData)
	})

	t.Run("handles initialize message type", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)
		cm.SetServerVersion("1.0.0")
		cm.SetAgentTypesProvider(&mockAgentTypesProvider{agentTypes: []AgentTypeInfo{}})

		runnerID := int64(102)
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
		initJSON, _ := json.Marshal(initParams)

		msg := RunnerMessage{
			Type: MsgTypeInitialize,
			Data: initJSON,
		}
		msgData, _ := json.Marshal(msg)

		cm.HandleMessage(runnerID, websocket.TextMessage, msgData)

		// Verify initialize_result was sent
		select {
		case data := <-conn.Send:
			var respMsg RunnerMessage
			json.Unmarshal(data, &respMsg)
			assert.Equal(t, MsgTypeInitializeResult, respMsg.Type)
		case <-time.After(time.Second):
			t.Fatal("expected initialize_result message")
		}
	})

	t.Run("handles initialized message type", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		callbackCalled := false
		cm.SetInitializedCallback(func(runnerID int64, availableAgents []string) {
			callbackCalled = true
		})

		runnerID := int64(103)
		conn := &RunnerConnection{
			RunnerID: runnerID,
			Conn:     &websocket.Conn{},
			Send:     make(chan []byte, 256),
		}
		shard := cm.getShard(runnerID)
		shard.mu.Lock()
		shard.connections[runnerID] = conn
		shard.mu.Unlock()

		initdParams := InitializedParams{
			AvailableAgents: []string{"claude-code"},
		}
		initdJSON, _ := json.Marshal(initdParams)

		msg := RunnerMessage{
			Type: MsgTypeInitialized,
			Data: initdJSON,
		}
		msgData, _ := json.Marshal(msg)

		cm.HandleMessage(runnerID, websocket.TextMessage, msgData)

		assert.True(t, callbackCalled)
	})

	t.Run("handles terminal_output message type", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		var receivedData *TerminalOutputData
		cm.SetTerminalOutputCallback(func(runnerID int64, data *TerminalOutputData) {
			receivedData = data
		})

		toData := TerminalOutputData{
			PodKey: "pod-1",
			Data:   []byte("hello world"),
		}
		toJSON, _ := json.Marshal(toData)

		msg := RunnerMessage{
			Type: MsgTypeTerminalOutput,
			Data: toJSON,
		}
		msgData, _ := json.Marshal(msg)

		cm.HandleMessage(100, websocket.TextMessage, msgData)

		require.NotNil(t, receivedData)
		assert.Equal(t, "pod-1", receivedData.PodKey)
		assert.Equal(t, []byte("hello world"), receivedData.Data)
	})

	t.Run("handles pty_resized message type", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		var receivedData *PtyResizedData
		cm.SetPtyResizedCallback(func(runnerID int64, data *PtyResizedData) {
			receivedData = data
		})

		prData := PtyResizedData{
			PodKey: "pod-1",
			Cols:   120,
			Rows:   40,
		}
		prJSON, _ := json.Marshal(prData)

		msg := RunnerMessage{
			Type: MsgTypePtyResized,
			Data: prJSON,
		}
		msgData, _ := json.Marshal(msg)

		cm.HandleMessage(100, websocket.TextMessage, msgData)

		require.NotNil(t, receivedData)
		assert.Equal(t, "pod-1", receivedData.PodKey)
		assert.Equal(t, 120, receivedData.Cols)
		assert.Equal(t, 40, receivedData.Rows)
	})

	t.Run("handles pod_created message type", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		var receivedData *PodCreatedData
		cm.SetPodCreatedCallback(func(runnerID int64, data *PodCreatedData) {
			receivedData = data
		})

		pcData := PodCreatedData{
			PodKey: "pod-created-1",
			Pid:    12345,
			Cols:   80,
			Rows:   24,
		}
		pcJSON, _ := json.Marshal(pcData)

		msg := RunnerMessage{
			Type: MsgTypePodCreated,
			Data: pcJSON,
		}
		msgData, _ := json.Marshal(msg)

		cm.HandleMessage(100, websocket.TextMessage, msgData)

		require.NotNil(t, receivedData)
		assert.Equal(t, "pod-created-1", receivedData.PodKey)
		assert.Equal(t, 12345, receivedData.Pid)
	})

	t.Run("handles pod_terminated message type", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		var receivedData *PodTerminatedData
		cm.SetPodTerminatedCallback(func(runnerID int64, data *PodTerminatedData) {
			receivedData = data
		})

		ptData := PodTerminatedData{
			PodKey:   "pod-terminated-1",
			ExitCode: 0,
		}
		ptJSON, _ := json.Marshal(ptData)

		msg := RunnerMessage{
			Type: MsgTypePodTerminated,
			Data: ptJSON,
		}
		msgData, _ := json.Marshal(msg)

		cm.HandleMessage(100, websocket.TextMessage, msgData)

		require.NotNil(t, receivedData)
		assert.Equal(t, "pod-terminated-1", receivedData.PodKey)
		assert.Equal(t, 0, receivedData.ExitCode)
	})

	t.Run("handles agent_status message type", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		var receivedData *AgentStatusData
		cm.SetAgentStatusCallback(func(runnerID int64, data *AgentStatusData) {
			receivedData = data
		})

		asData := AgentStatusData{
			PodKey: "pod-agent-1",
			Status: "running",
			Pid:    9999,
		}
		asJSON, _ := json.Marshal(asData)

		msg := RunnerMessage{
			Type: MsgTypeAgentStatus,
			Data: asJSON,
		}
		msgData, _ := json.Marshal(msg)

		cm.HandleMessage(100, websocket.TextMessage, msgData)

		require.NotNil(t, receivedData)
		assert.Equal(t, "pod-agent-1", receivedData.PodKey)
		assert.Equal(t, "running", receivedData.Status)
	})
}

func TestHandleMessageParsing(t *testing.T) {
	t.Run("handles pod_created message", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		var receivedData *PodCreatedData
		cm.SetPodCreatedCallback(func(runnerID int64, data *PodCreatedData) {
			receivedData = data
		})

		pcData := PodCreatedData{
			PodKey: "test-pod-123",
			Pid:    12345,
			Cols:   80,
			Rows:   24,
		}
		data, _ := json.Marshal(pcData)

		cm.handlePodCreatedMessage(123, data)

		require.NotNil(t, receivedData)
		assert.Equal(t, "test-pod-123", receivedData.PodKey)
		assert.Equal(t, 12345, receivedData.Pid)
	})

	t.Run("handles pod_terminated message", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		var receivedData *PodTerminatedData
		cm.SetPodTerminatedCallback(func(runnerID int64, data *PodTerminatedData) {
			receivedData = data
		})

		ptData := PodTerminatedData{
			PodKey:   "test-pod-456",
			ExitCode: 0,
		}
		data, _ := json.Marshal(ptData)

		cm.handlePodTerminatedMessage(456, data)

		require.NotNil(t, receivedData)
		assert.Equal(t, "test-pod-456", receivedData.PodKey)
		assert.Equal(t, 0, receivedData.ExitCode)
	})

	t.Run("handles agent_status message", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		var receivedData *AgentStatusData
		cm.SetAgentStatusCallback(func(runnerID int64, data *AgentStatusData) {
			receivedData = data
		})

		asData := AgentStatusData{
			PodKey: "test-pod",
			Status: "running",
			Pid:    9999,
		}
		data, _ := json.Marshal(asData)

		cm.handleAgentStatusMessage(789, data)

		require.NotNil(t, receivedData)
		assert.Equal(t, "test-pod", receivedData.PodKey)
		assert.Equal(t, "running", receivedData.Status)
	})

	t.Run("logs error on invalid JSON", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		called := false
		cm.SetPodCreatedCallback(func(runnerID int64, data *PodCreatedData) {
			called = true
		})

		// Should not panic and should not call callback
		cm.handlePodCreatedMessage(123, json.RawMessage(`{invalid}`))
		assert.False(t, called)
	})
}

func TestHandleHeartbeatMessage(t *testing.T) {
	t.Run("updates heartbeat and calls callback", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		var receivedRunnerID int64
		var receivedData *HeartbeatData
		cm.SetHeartbeatCallback(func(runnerID int64, data *HeartbeatData) {
			receivedRunnerID = runnerID
			receivedData = data
		})

		runnerID := int64(200)
		conn := &RunnerConnection{
			RunnerID: runnerID,
			Conn:     &websocket.Conn{},
			Send:     make(chan []byte, 256),
			LastPing: time.Now().Add(-time.Hour), // Old heartbeat
		}
		shard := cm.getShard(runnerID)
		shard.mu.Lock()
		shard.connections[runnerID] = conn
		shard.mu.Unlock()

		hbData := HeartbeatData{
			Pods: []HeartbeatPod{{PodKey: "pod-1"}},
		}
		data, _ := json.Marshal(hbData)

		cm.handleHeartbeatMessage(runnerID, data)

		// Verify callback was called
		assert.Equal(t, runnerID, receivedRunnerID)
		require.NotNil(t, receivedData)
		assert.Len(t, receivedData.Pods, 1)

		// Verify heartbeat was updated
		conn.mu.Lock()
		assert.True(t, conn.LastPing.After(time.Now().Add(-time.Minute)))
		conn.mu.Unlock()
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		callbackCalled := false
		cm.SetHeartbeatCallback(func(runnerID int64, data *HeartbeatData) {
			callbackCalled = true
		})

		cm.handleHeartbeatMessage(100, json.RawMessage(`{invalid}`))
		assert.False(t, callbackCalled)
	})

	t.Run("handles no callback set", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)
		// No callback set

		runnerID := int64(201)
		conn := &RunnerConnection{
			RunnerID: runnerID,
			Conn:     &websocket.Conn{},
			Send:     make(chan []byte, 256),
			LastPing: time.Now().Add(-time.Hour),
		}
		shard := cm.getShard(runnerID)
		shard.mu.Lock()
		shard.connections[runnerID] = conn
		shard.mu.Unlock()

		hbData := HeartbeatData{Pods: []HeartbeatPod{}}
		data, _ := json.Marshal(hbData)

		// Should not panic even without callback
		cm.handleHeartbeatMessage(runnerID, data)
	})
}

func TestHandleTerminalOutputMessage(t *testing.T) {
	t.Run("calls callback with data", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		var receivedRunnerID int64
		var receivedData *TerminalOutputData
		cm.SetTerminalOutputCallback(func(runnerID int64, data *TerminalOutputData) {
			receivedRunnerID = runnerID
			receivedData = data
		})

		toData := TerminalOutputData{
			PodKey: "test-pod",
			Data:   []byte("output line 1\noutput line 2\n"),
		}
		data, _ := json.Marshal(toData)

		cm.handleTerminalOutputMessage(300, data)

		assert.Equal(t, int64(300), receivedRunnerID)
		require.NotNil(t, receivedData)
		assert.Equal(t, "test-pod", receivedData.PodKey)
		assert.Equal(t, []byte("output line 1\noutput line 2\n"), receivedData.Data)
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		callbackCalled := false
		cm.SetTerminalOutputCallback(func(runnerID int64, data *TerminalOutputData) {
			callbackCalled = true
		})

		cm.handleTerminalOutputMessage(300, json.RawMessage(`{invalid}`))
		assert.False(t, callbackCalled)
	})

	t.Run("handles no callback set", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)
		// No callback set

		toData := TerminalOutputData{PodKey: "test", Data: []byte("data")}
		data, _ := json.Marshal(toData)

		// Should not panic
		cm.handleTerminalOutputMessage(300, data)
	})
}

func TestHandlePtyResizedMessage(t *testing.T) {
	t.Run("calls callback with data", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		var receivedRunnerID int64
		var receivedData *PtyResizedData
		cm.SetPtyResizedCallback(func(runnerID int64, data *PtyResizedData) {
			receivedRunnerID = runnerID
			receivedData = data
		})

		prData := PtyResizedData{
			PodKey: "resize-pod",
			Cols:   100,
			Rows:   30,
		}
		data, _ := json.Marshal(prData)

		cm.handlePtyResizedMessage(400, data)

		assert.Equal(t, int64(400), receivedRunnerID)
		require.NotNil(t, receivedData)
		assert.Equal(t, "resize-pod", receivedData.PodKey)
		assert.Equal(t, 100, receivedData.Cols)
		assert.Equal(t, 30, receivedData.Rows)
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)

		callbackCalled := false
		cm.SetPtyResizedCallback(func(runnerID int64, data *PtyResizedData) {
			callbackCalled = true
		})

		cm.handlePtyResizedMessage(400, json.RawMessage(`{invalid}`))
		assert.False(t, callbackCalled)
	})

	t.Run("handles no callback set", func(t *testing.T) {
		logger := newTestLogger()
		cm := NewConnectionManager(logger)
		// No callback set

		prData := PtyResizedData{PodKey: "test", Cols: 80, Rows: 24}
		data, _ := json.Marshal(prData)

		// Should not panic
		cm.handlePtyResizedMessage(400, data)
	})
}

func TestHandlePodTerminatedMessage_InvalidJSON(t *testing.T) {
	logger := newTestLogger()
	cm := NewConnectionManager(logger)

	callbackCalled := false
	cm.SetPodTerminatedCallback(func(runnerID int64, data *PodTerminatedData) {
		callbackCalled = true
	})

	cm.handlePodTerminatedMessage(500, json.RawMessage(`{invalid}`))
	assert.False(t, callbackCalled)
}

func TestHandleAgentStatusMessage_InvalidJSON(t *testing.T) {
	logger := newTestLogger()
	cm := NewConnectionManager(logger)

	callbackCalled := false
	cm.SetAgentStatusCallback(func(runnerID int64, data *AgentStatusData) {
		callbackCalled = true
	})

	cm.handleAgentStatusMessage(600, json.RawMessage(`{invalid}`))
	assert.False(t, callbackCalled)
}

func TestConnectionClose(t *testing.T) {
	t.Run("close is idempotent", func(t *testing.T) {
		conn := &RunnerConnection{
			RunnerID: 123,
			Conn:     nil,
			Send:     make(chan []byte, 10),
		}

		// Should not panic on multiple closes
		conn.Close()
		conn.Close()
		conn.Close()
	})

	t.Run("closes channel exactly once", func(t *testing.T) {
		conn := &RunnerConnection{
			RunnerID: 456,
			Conn:     nil,
			Send:     make(chan []byte, 10),
		}

		// Send a message before close
		conn.Send <- []byte("test")

		conn.Close()

		// First read should return the buffered message
		msg, ok := <-conn.Send
		assert.True(t, ok, "should return buffered message")
		assert.Equal(t, []byte("test"), msg)

		// Second read should indicate channel is closed
		_, ok = <-conn.Send
		assert.False(t, ok, "channel should be closed after draining buffer")
	})
}
