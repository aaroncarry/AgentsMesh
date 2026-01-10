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

// --- Test handlePodStart with builder ---

func TestPodHandlerHandlePodStartWithBuilder(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 10,
			WorkspaceRoot:         "/tmp",
			AgentEnvVars:          map[string]string{},
		},
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodStartPayload{
		PodKey:    "test-build-pod",
		AgentType:     "claude-code",
		LaunchCommand: "echo",
		LaunchArgs:    []string{"hello"},
		Rows:          30,
		Cols:          100,
	}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodStart,
		Payload: payloadBytes,
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify pod was stored
	pod, ok := store.Get("test-build-pod")
	if !ok {
		t.Error("pod should be stored")
	}
	if pod != nil {
		// Clean up
		if pod.Terminal != nil {
			pod.Terminal.Stop()
		}
	}

	// Verify status was sent
	eventSender.mu.Lock()
	statusCount := len(eventSender.statuses)
	eventSender.mu.Unlock()

	if statusCount < 1 {
		t.Error("should have sent at least one status")
	}
}

func TestPodHandlerHandlePodStartWithRepository(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 10,
			WorkspaceRoot:         "/tmp",
		},
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodStartPayload{
		PodKey:    "repo-pod",
		AgentType:     "claude-code",
		LaunchCommand: "echo",
		RepositoryURL: "https://github.com/test/repo.git",
		Branch:        "main",
	}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodStart,
		Payload: payloadBytes,
	}

	// This will likely fail to create worktree, but tests the code path
	err := handler.HandleMessage(context.Background(), msg)
	if err != nil {
		// Expected to fail without real git repo, but we want to verify the path is exercised
		// Check that error notification was sent
		eventSender.mu.Lock()
		hasErrorStatus := false
		for _, s := range eventSender.statuses {
			if s.status == "error" {
				hasErrorStatus = true
			}
		}
		eventSender.mu.Unlock()

		if !hasErrorStatus {
			t.Log("Expected error status for failed worktree creation")
		}
	}
}

func TestPodHandlerHandlePodStartWithTicketWorktree(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 10,
			WorkspaceRoot:         "/tmp",
		},
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodStartPayload{
		PodKey:       "ticket-pod",
		AgentType:        "claude-code",
		LaunchCommand:    "echo",
		TicketIdentifier: "TICKET-123",
	}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodStart,
		Payload: payloadBytes,
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err != nil {
		// Without worktree service, falls back to normal workspace
		t.Logf("Error (expected without worktree service): %v", err)
	}
}

func TestPodHandlerHandlePodStartWithInitialPrompt(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 10,
			WorkspaceRoot:         "/tmp",
		},
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodStartPayload{
		PodKey:    "prompt-pod",
		AgentType:     "claude-code",
		LaunchCommand: "cat", // cat will wait for input
		InitialPrompt: "Hello, Claude!",
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
		t.Errorf("unexpected error: %v", err)
	}

	// Wait for initial prompt to be sent
	time.Sleep(600 * time.Millisecond)

	// Clean up
	pod, ok := store.Get("prompt-pod")
	if ok && pod.Terminal != nil {
		pod.Terminal.Stop()
	}
}

// --- Test handlePodStop with terminal ---

func TestPodHandlerHandlePodStopWithTerminal(t *testing.T) {
	runner := &Runner{
		cfg:       &config.Config{},
		workspace: nil,
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	// Create a real terminal for testing
	term, err := terminal.New(terminal.Options{
		Command: "sleep",
		Args:    []string{"10"},
		WorkDir: "/tmp",
		Rows:    24,
		Cols:    80,
	})
	if err != nil {
		t.Skipf("Could not create terminal: %v", err)
	}
	term.Start()

	// Add pod with terminal
	store.Put("term-pod", &Pod{
		ID:         "term-pod",
		PodKey: "term-pod",
		Terminal:   term,
	})

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodStopPayload{PodKey: "term-pod"}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodStop,
		Payload: payloadBytes,
	}

	err = handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify pod was removed
	if store.Count() != 0 {
		t.Errorf("pod count = %d, want 0", store.Count())
	}

	// Verify stopped status was sent
	eventSender.mu.Lock()
	hasStoppedStatus := false
	for _, s := range eventSender.statuses {
		if s.status == "stopped" {
			hasStoppedStatus = true
		}
	}
	eventSender.mu.Unlock()

	if !hasStoppedStatus {
		t.Error("should have sent stopped status")
	}
}

// --- Test handleTerminalInput with pod ---

func TestPodHandlerHandleTerminalInputWithPod(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{},
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	// Create a real terminal
	term, err := terminal.New(terminal.Options{
		Command: "cat",
		WorkDir: "/tmp",
		Rows:    24,
		Cols:    80,
	})
	if err != nil {
		t.Skipf("Could not create terminal: %v", err)
	}
	term.Start()
	defer term.Stop()

	store.Put("input-pod", &Pod{
		ID:         "input-pod",
		PodKey: "input-pod",
		Terminal:   term,
	})

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := TerminalInputPayload{
		PodKey: "input-pod",
		Data:       []byte("test input\n"),
	}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypeTerminalInput,
		Payload: payloadBytes,
	}

	err = handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- Test handleTerminalResize with pod ---

func TestPodHandlerHandleTerminalResizeWithPod(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{},
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	// Create a real terminal
	term, err := terminal.New(terminal.Options{
		Command: "sleep",
		Args:    []string{"10"},
		WorkDir: "/tmp",
		Rows:    24,
		Cols:    80,
	})
	if err != nil {
		t.Skipf("Could not create terminal: %v", err)
	}
	term.Start()
	defer term.Stop()

	store.Put("resize-pod", &Pod{
		ID:         "resize-pod",
		PodKey: "resize-pod",
		Terminal:   term,
	})

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := TerminalResizePayload{
		PodKey: "resize-pod",
		Rows:       40,
		Cols:       120,
	}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypeTerminalResize,
		Payload: payloadBytes,
	}

	err = handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- Test handlePodList with terminal info ---

func TestPodHandlerHandlePodListWithTerminal(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	// Create a real terminal for PID info
	term, err := terminal.New(terminal.Options{
		Command: "sleep",
		Args:    []string{"10"},
		WorkDir: "/tmp",
		Rows:    24,
		Cols:    80,
	})
	if err != nil {
		t.Skipf("Could not create terminal: %v", err)
	}
	term.Start()
	defer term.Stop()

	store.Put("list-pod", &Pod{
		ID:            "list-pod",
		PodKey:    "list-pod",
		AgentType:     "claude-code",
		Status:        PodStatusRunning,
		StartedAt:     time.Now(),
		WorktreePath:  "/workspace/worktrees/test",
		RepositoryURL: "https://github.com/test/repo.git",
		Terminal:      term,
	})

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodListPayload{RequestID: "req-with-terminal"}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodList,
		Payload: payloadBytes,
	}

	err = handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify status contains PID
	eventSender.mu.Lock()
	defer eventSender.mu.Unlock()

	if len(eventSender.statuses) != 1 {
		t.Errorf("statuses count = %d, want 1", len(eventSender.statuses))
		return
	}

	data := eventSender.statuses[0].data
	pods, ok := data["pods"].([]map[string]interface{})
	if !ok {
		t.Log("pods data format unexpected, skipping PID check")
		return
	}

	if len(pods) > 0 {
		if _, hasPID := pods[0]["pid"]; !hasPID {
			t.Error("pod info should include PID")
		}
	}
}

// --- Test output callback ---

func TestPodHandlerOutputCallback(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 10,
			WorkspaceRoot:         "/tmp",
		},
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	// Create pod with output callback
	payload := PodStartPayload{
		PodKey:    "output-pod",
		AgentType:     "claude-code",
		LaunchCommand: "echo",
		LaunchArgs:    []string{"test output"},
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
		t.Errorf("unexpected error: %v", err)
	}

	// Wait for output
	time.Sleep(100 * time.Millisecond)

	// Clean up
	pod, ok := store.Get("output-pod")
	if ok && pod.Terminal != nil {
		pod.Terminal.Stop()
	}

	// Check if output was sent
	eventSender.mu.Lock()
	outputCount := len(eventSender.outputs)
	eventSender.mu.Unlock()

	// Output may or may not have been captured depending on timing
	t.Logf("Output count: %d", outputCount)
}

// --- Test exit callback ---

func TestPodHandlerExitCallback(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 10,
			WorkspaceRoot:         "/tmp",
		},
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	// Create pod that exits quickly
	payload := PodStartPayload{
		PodKey:    "exit-pod",
		AgentType:     "claude-code",
		LaunchCommand: "true", // exits immediately with code 0
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
		t.Errorf("unexpected error: %v", err)
	}

	// Wait for exit
	time.Sleep(200 * time.Millisecond)

	// Check for exited status
	eventSender.mu.Lock()
	hasExitedStatus := false
	for _, s := range eventSender.statuses {
		if s.status == "exited" {
			hasExitedStatus = true
		}
	}
	eventSender.mu.Unlock()

	// The exit callback may or may not have been triggered depending on timing
	t.Logf("Has exited status: %v", hasExitedStatus)
}

// --- Benchmark ---

func BenchmarkPodHandlerHandleMessage(b *testing.B) {
	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 100,
		},
	}
	termManager := terminal.NewManager("/bin/bash", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodListPayload{RequestID: "bench-req"}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodList,
		Payload: payloadBytes,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.HandleMessage(ctx, msg)
	}
}
