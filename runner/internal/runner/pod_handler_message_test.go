package runner

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/anthropics/agentmesh/runner/internal/client"
	"github.com/anthropics/agentmesh/runner/internal/config"
	"github.com/anthropics/agentmesh/runner/internal/terminal"
)

// Tests for HandleMessage routing and JSON parsing

func TestPodHandlerHandleMessageUnknownType(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	msg := &client.Message{
		Type:    "unknown_type",
		Payload: []byte("{}"),
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("unexpected error for unknown type: %v", err)
	}
}

// --- PodStop Tests ---

func TestPodHandlerHandlePodStopNotFound(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodStopPayload{PodKey: "nonexistent"}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodStop,
		Payload: payloadBytes,
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err == nil {
		t.Error("expected error for nonexistent pod")
	}
}

func TestPodHandlerHandlePodStopWithWorktree(t *testing.T) {
	runner := &Runner{
		cfg:       &config.Config{},
		workspace: nil, // No workspace manager
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	// Add pod with worktree
	store.Put("pod-1", &Pod{
		ID:           "pod-1",
		PodKey:   "pod-1",
		WorktreePath: "/workspace/worktrees/pod-1",
		Terminal:     nil,
	})

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodStopPayload{PodKey: "pod-1"}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodStop,
		Payload: payloadBytes,
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Pod should be removed
	if store.Count() != 0 {
		t.Errorf("pod count = %d, want 0", store.Count())
	}

	// Check status sent
	eventSender.mu.Lock()
	defer eventSender.mu.Unlock()

	if len(eventSender.statuses) != 1 {
		t.Errorf("statuses count = %d, want 1", len(eventSender.statuses))
	}
	if eventSender.statuses[0].status != "stopped" {
		t.Errorf("status = %v, want stopped", eventSender.statuses[0].status)
	}
}

// --- TerminalInput Tests ---

func TestPodHandlerHandleTerminalInputNotFound(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := TerminalInputPayload{PodKey: "nonexistent", Data: []byte("test")}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypeTerminalInput,
		Payload: payloadBytes,
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err == nil {
		t.Error("expected error for nonexistent pod")
	}
}

// --- TerminalResize Tests ---

func TestPodHandlerHandleTerminalResizeNotFound(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := TerminalResizePayload{PodKey: "nonexistent", Rows: 40, Cols: 120}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypeTerminalResize,
		Payload: payloadBytes,
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err == nil {
		t.Error("expected error for nonexistent pod")
	}
}

// --- PodList Tests ---

func TestPodHandlerHandlePodList(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	// Add some pods
	store.Put("pod-1", &Pod{
		ID:         "pod-1",
		PodKey: "pod-1",
		AgentType:  "claude-code",
		Status:     PodStatusRunning,
		StartedAt:  time.Now(),
	})

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodListPayload{RequestID: "req-123"}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodList,
		Payload: payloadBytes,
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check that status was sent
	eventSender.mu.Lock()
	defer eventSender.mu.Unlock()

	if len(eventSender.statuses) != 1 {
		t.Errorf("statuses count: got %v, want 1", len(eventSender.statuses))
	}

	if eventSender.statuses[0].status != "pod_list" {
		t.Errorf("status: got %v, want pod_list", eventSender.statuses[0].status)
	}
}

func TestPodHandlerHandlePodListWithMultiplePods(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	// Add multiple pods
	store.Put("pod-1", &Pod{
		ID:         "pod-1",
		PodKey: "pod-1",
		AgentType:  "claude-code",
		Status:     PodStatusRunning,
		StartedAt:  time.Now(),
	})
	store.Put("pod-2", &Pod{
		ID:            "pod-2",
		PodKey:    "pod-2",
		AgentType:     "aider",
		Status:        PodStatusRunning,
		StartedAt:     time.Now(),
		WorktreePath:  "/workspace/worktrees/pod-2",
		RepositoryURL: "https://github.com/test/repo.git",
	})

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodListPayload{RequestID: "req-456"}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodList,
		Payload: payloadBytes,
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	eventSender.mu.Lock()
	defer eventSender.mu.Unlock()

	if len(eventSender.statuses) != 1 {
		t.Errorf("statuses count: got %v, want 1", len(eventSender.statuses))
	}
}

func TestPodHandlerHandlePodListEmpty(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodListPayload{RequestID: "req-empty"}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodList,
		Payload: payloadBytes,
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	eventSender.mu.Lock()
	defer eventSender.mu.Unlock()

	if len(eventSender.statuses) != 1 {
		t.Errorf("statuses count = %d, want 1", len(eventSender.statuses))
	}
}

// --- Invalid JSON Tests ---

func TestPodHandlerHandlePodStartInvalidJSON(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	msg := &client.Message{
		Type:    client.MessageTypePodStart,
		Payload: []byte("invalid json"),
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestPodHandlerHandlePodStopInvalidJSON(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	msg := &client.Message{
		Type:    client.MessageTypePodStop,
		Payload: []byte("invalid json"),
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestPodHandlerHandleTerminalInputInvalidJSON(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	msg := &client.Message{
		Type:    client.MessageTypeTerminalInput,
		Payload: []byte("invalid json"),
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestPodHandlerHandleTerminalResizeInvalidJSON(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	msg := &client.Message{
		Type:    client.MessageTypeTerminalResize,
		Payload: []byte("invalid json"),
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestPodHandlerHandlePodListInvalidJSON(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	msg := &client.Message{
		Type:    client.MessageTypePodList,
		Payload: []byte("invalid json"),
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
