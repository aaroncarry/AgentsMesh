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

// --- Additional tests for PodHandler ---

func TestPodHandlerHandleAllMessageTypes(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 10,
		},
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	tests := []struct {
		name    string
		msgType string
		payload interface{}
		wantErr bool
	}{
		{
			name:    "heartbeat - unknown type",
			msgType: client.MessageTypeHeartbeat,
			payload: map[string]string{},
			wantErr: false,
		},
		{
			name:    "error - unknown type",
			msgType: client.MessageTypeError,
			payload: map[string]string{"error": "test"},
			wantErr: false,
		},
		{
			name:    "custom unknown type",
			msgType: "custom.unknown",
			payload: map[string]string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, _ := json.Marshal(tt.payload)
			msg := &client.Message{
				Type:    tt.msgType,
				Payload: payloadBytes,
			}

			err := handler.HandleMessage(context.Background(), msg)
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestPodHandlerHandlePodListWithPID(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{},
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	// Add pod (Terminal is nil in this test)
	store.Put("pod-1", &Pod{
		ID:            "pod-1",
		PodKey:    "pod-1",
		AgentType:     "claude-code",
		Status:        PodStatusRunning,
		StartedAt:     time.Now(),
		WorktreePath:  "/test/path",
		RepositoryURL: "https://github.com/test/repo.git",
		Terminal:      nil, // nil terminal
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

	// Verify status was sent
	eventSender.mu.Lock()
	defer eventSender.mu.Unlock()

	if len(eventSender.statuses) != 1 {
		t.Errorf("statuses count = %d, want 1", len(eventSender.statuses))
	}

	if eventSender.statuses[0].status != "pod_list" {
		t.Errorf("status = %v, want pod_list", eventSender.statuses[0].status)
	}
}

func TestPodHandlerHandlePodExitUpdatesStatus(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{},
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	// Add pod
	pod := &Pod{
		ID:     "pod-1",
		Status: PodStatusRunning,
	}
	store.Put("pod-1", pod)

	handler := NewPodHandler(runner, termManager, eventSender, store)

	// Handle exit
	handler.handlePodExit("pod-1", 42)

	// Pod should be deleted from store
	_, ok := store.Get("pod-1")
	if ok {
		t.Error("pod should be removed from store")
	}

	// Event should be sent
	eventSender.mu.Lock()
	defer eventSender.mu.Unlock()

	if len(eventSender.statuses) != 1 {
		t.Errorf("statuses count = %d, want 1", len(eventSender.statuses))
	}

	if eventSender.statuses[0].status != "exited" {
		t.Errorf("status = %v, want exited", eventSender.statuses[0].status)
	}

	exitCode := eventSender.statuses[0].data["exit_code"]
	if exitCode != 42 {
		t.Errorf("exit_code = %v, want 42", exitCode)
	}
}

func TestPodHandlerSendPodErrorFormat(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{},
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	// Send error
	handler.sendPodError("pod-1", "test error message")

	// Verify
	eventSender.mu.Lock()
	defer eventSender.mu.Unlock()

	if len(eventSender.statuses) != 1 {
		t.Errorf("statuses count = %d, want 1", len(eventSender.statuses))
	}

	status := eventSender.statuses[0]
	if status.podKey != "pod-1" {
		t.Errorf("podKey = %v, want pod-1", status.podKey)
	}

	if status.status != "error" {
		t.Errorf("status = %v, want error", status.status)
	}

	if status.data["error"] != "test error message" {
		t.Errorf("error = %v, want 'test error message'", status.data["error"])
	}
}

// --- InMemoryPodStore tests ---

func TestInMemoryPodStoreOperations(t *testing.T) {
	store := NewInMemoryPodStore()

	// Test Put and Get
	pod := &Pod{ID: "pod-1", Status: PodStatusRunning}
	store.Put("pod-1", pod)

	got, ok := store.Get("pod-1")
	if !ok {
		t.Error("pod should exist")
	}
	if got.ID != "pod-1" {
		t.Errorf("ID = %v, want pod-1", got.ID)
	}

	// Test Count
	if store.Count() != 1 {
		t.Errorf("Count = %d, want 1", store.Count())
	}

	// Test All
	all := store.All()
	if len(all) != 1 {
		t.Errorf("All() length = %d, want 1", len(all))
	}

	// Test Delete
	deleted := store.Delete("pod-1")
	if deleted == nil {
		t.Error("deleted should not be nil")
	}
	if deleted.ID != "pod-1" {
		t.Errorf("deleted ID = %v, want pod-1", deleted.ID)
	}

	// Test Get after Delete
	_, ok = store.Get("pod-1")
	if ok {
		t.Error("pod should not exist after delete")
	}

	// Test Count after Delete
	if store.Count() != 0 {
		t.Errorf("Count after delete = %d, want 0", store.Count())
	}
}

func TestInMemoryPodStoreDeleteNonExistent(t *testing.T) {
	store := NewInMemoryPodStore()

	deleted := store.Delete("nonexistent")
	if deleted != nil {
		t.Error("delete of nonexistent should return nil")
	}
}

func TestInMemoryPodStoreAllEmpty(t *testing.T) {
	store := NewInMemoryPodStore()

	all := store.All()
	if len(all) != 0 {
		t.Errorf("All() on empty store = %d, want 0", len(all))
	}
}

func TestInMemoryPodStoreMultiplePods(t *testing.T) {
	store := NewInMemoryPodStore()

	// Add multiple pods
	for i := 1; i <= 5; i++ {
		store.Put("pod-"+string(rune('0'+i)), &Pod{
			ID:     "pod-" + string(rune('0'+i)),
			Status: PodStatusRunning,
		})
	}

	if store.Count() != 5 {
		t.Errorf("Count = %d, want 5", store.Count())
	}

	all := store.All()
	if len(all) != 5 {
		t.Errorf("All() length = %d, want 5", len(all))
	}
}

func TestInMemoryPodStoreOverwrite(t *testing.T) {
	store := NewInMemoryPodStore()

	// Add pod
	store.Put("pod-1", &Pod{ID: "pod-1", Status: PodStatusRunning})

	// Overwrite
	store.Put("pod-1", &Pod{ID: "pod-1", Status: PodStatusStopped})

	got, ok := store.Get("pod-1")
	if !ok {
		t.Error("pod should exist")
	}
	if got.Status != PodStatusStopped {
		t.Errorf("Status = %v, want stopped", got.Status)
	}

	// Count should still be 1
	if store.Count() != 1 {
		t.Errorf("Count after overwrite = %d, want 1", store.Count())
	}
}
