package runner

import (
	"testing"

	"github.com/anthropics/agentmesh/runner/internal/terminal"
)

// Basic unit tests for PodHandler creation and helper methods

func TestNewPodHandler(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	if handler == nil {
		t.Fatal("NewPodHandler returned nil")
	}

	if handler.runner != runner {
		t.Error("runner should be set")
	}

	if handler.termManager != termManager {
		t.Error("termManager should be set")
	}

	if handler.eventSender != eventSender {
		t.Error("eventSender should be set")
	}

	if handler.podStore != store {
		t.Error("podStore should be set")
	}
}

func TestPodHandlerHandlePodExit(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	// Add pod
	store.Put("pod-1", &Pod{
		ID:         "pod-1",
		PodKey: "pod-1",
		Status:     PodStatusRunning,
	})

	handler := NewPodHandler(runner, termManager, eventSender, store)

	handler.handlePodExit("pod-1", 0)

	// Pod should be removed
	_, ok := store.Get("pod-1")
	if ok {
		t.Error("pod should be removed after exit")
	}

	// Check that exit status was sent
	eventSender.mu.Lock()
	defer eventSender.mu.Unlock()

	if len(eventSender.statuses) != 1 {
		t.Errorf("statuses count: got %v, want 1", len(eventSender.statuses))
	}

	if eventSender.statuses[0].status != "exited" {
		t.Errorf("status: got %v, want exited", eventSender.statuses[0].status)
	}
}

func TestPodHandlerHandlePodExitNonExistent(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	// Handle exit for non-existent pod
	handler.handlePodExit("nonexistent", 0)

	eventSender.mu.Lock()
	defer eventSender.mu.Unlock()

	// Should still send exited status
	if len(eventSender.statuses) != 1 {
		t.Errorf("statuses count = %d, want 1", len(eventSender.statuses))
	}
	if eventSender.statuses[0].status != "exited" {
		t.Errorf("status = %v, want exited", eventSender.statuses[0].status)
	}
}

func TestPodHandlerSendPodError(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	handler.sendPodError("pod-1", "test error")

	eventSender.mu.Lock()
	defer eventSender.mu.Unlock()

	if len(eventSender.statuses) != 1 {
		t.Errorf("statuses count: got %v, want 1", len(eventSender.statuses))
	}

	if eventSender.statuses[0].status != "error" {
		t.Errorf("status: got %v, want error", eventSender.statuses[0].status)
	}

	if eventSender.statuses[0].data["error"] != "test error" {
		t.Errorf("error message: got %v, want test error", eventSender.statuses[0].data["error"])
	}
}
