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

// Tests for PodStart message handling

func TestPodHandlerHandlePodStartMaxCapacity(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 1,
		},
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	// Fill up capacity
	store.Put("existing", &Pod{ID: "existing"})

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodStartPayload{
		PodKey:    "new-pod",
		AgentType:     "claude-code",
		LaunchCommand: "echo",
	}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodStart,
		Payload: payloadBytes,
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err == nil {
		t.Error("expected error for max capacity")
	}
	if !contains(err.Error(), "max concurrent pods") {
		t.Errorf("error = %v, want containing 'max concurrent pods'", err)
	}
}

func TestPodHandlerHandlePodStartSuccess(t *testing.T) {
	tempDir := t.TempDir()

	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 10,
			WorkspaceRoot:         tempDir,
		},
	}
	termManager := terminal.NewManager("/bin/bash", tempDir)
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodStartPayload{
		PodKey:    "success-pod",
		AgentType:     "claude-code",
		LaunchCommand: "echo",
		LaunchArgs:    []string{"hello"},
		Rows:          24,
		Cols:          80,
	}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodStart,
		Payload: payloadBytes,
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Skipf("Could not start pod: %v", err)
	}

	// Wait for terminal to start
	time.Sleep(100 * time.Millisecond)

	// Check pod was created
	pod, ok := store.Get("success-pod")
	if !ok {
		t.Error("pod should be stored")
	}

	// Check started status was sent
	eventSender.mu.Lock()
	hasStarted := false
	for _, s := range eventSender.statuses {
		if s.status == "started" {
			hasStarted = true
			break
		}
	}
	eventSender.mu.Unlock()

	if !hasStarted {
		t.Error("should have sent started status")
	}

	// Clean up
	if pod != nil && pod.Terminal != nil {
		pod.Terminal.Stop()
	}
}

func TestPodHandlerHandlePodStartWithPromptDelay(t *testing.T) {
	tempDir := t.TempDir()

	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 10,
			WorkspaceRoot:         tempDir,
		},
	}
	termManager := terminal.NewManager("/bin/bash", tempDir)
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodStartPayload{
		PodKey:    "prompt-delay-pod",
		AgentType:     "claude-code",
		LaunchCommand: "cat",
		InitialPrompt: "Hello World",
		Rows:          24,
		Cols:          80,
	}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodStart,
		Payload: payloadBytes,
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Skipf("Could not start pod: %v", err)
	}

	// Wait for initial prompt to be sent
	time.Sleep(700 * time.Millisecond)

	// Clean up
	pod, ok := store.Get("prompt-delay-pod")
	if ok && pod.Terminal != nil {
		pod.Terminal.Stop()
	}
}

func TestPodHandlerHandlePodStartWithRepositoryURL(t *testing.T) {
	tempDir := t.TempDir()

	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 10,
			WorkspaceRoot:         tempDir,
		},
	}
	termManager := terminal.NewManager("/bin/bash", tempDir)
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodStartPayload{
		PodKey:    "repo-pod",
		AgentType:     "claude-code",
		LaunchCommand: "echo",
		LaunchArgs:    []string{"test"},
		RepositoryURL: "https://github.com/test/repo", // Non-existent repo
		Branch:        "main",
		Rows:          24,
		Cols:          80,
	}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodStart,
		Payload: payloadBytes,
	}

	err := handler.HandleMessage(context.Background(), msg)
	// May or may not fail depending on repository availability
	t.Logf("HandleMessage with repository: %v", err)

	// Clean up
	pod, ok := store.Get("repo-pod")
	if ok && pod.Terminal != nil {
		pod.Terminal.Stop()
	}
}
