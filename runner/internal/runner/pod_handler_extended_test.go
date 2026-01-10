package runner

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/anthropics/agentmesh/runner/internal/client"
	"github.com/anthropics/agentmesh/runner/internal/config"
	"github.com/anthropics/agentmesh/runner/internal/terminal"
	"github.com/anthropics/agentmesh/runner/internal/workspace"
)

// errorEventSenderStatus represents a captured status event
type errorEventSenderStatus struct {
	podKey string
	status     string
	data       map[string]interface{}
}

// errorEventSender is a mock that returns errors for SendTerminalOutput
type errorEventSender struct {
	statuses    []errorEventSenderStatus
	outputError error
	mu          sync.Mutex
}

func newErrorEventSender(err error) *errorEventSender {
	return &errorEventSender{outputError: err}
}

func (m *errorEventSender) SendPodStatus(podKey, status string, data map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statuses = append(m.statuses, errorEventSenderStatus{podKey, status, data})
}

func (m *errorEventSender) SendTerminalOutput(podKey string, data []byte) error {
	return m.outputError
}

// --- Test handlePodStart with initial prompt ---

func TestPodHandlerStartWithInitialPrompt(t *testing.T) {
	tempDir := t.TempDir()
	ws, err := workspace.NewManager(tempDir, "")
	if err != nil {
		t.Skipf("Could not create workspace manager: %v", err)
	}

	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 10,
			WorkspaceRoot:         tempDir,
		},
		workspace: ws,
	}
	termManager := terminal.NewManager("/bin/sh", tempDir)
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodStartPayload{
		PodKey:    "prompt-pod",
		AgentType:     "claude-code",
		LaunchCommand: "cat",
		Rows:          24,
		Cols:          80,
		InitialPrompt: "Hello, world!",
	}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodStart,
		Payload: payloadBytes,
	}

	err = handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Wait for initial prompt to be sent (500ms + buffer)
	time.Sleep(700 * time.Millisecond)

	// Clean up
	pod, ok := store.Get("prompt-pod")
	if ok && pod.Terminal != nil {
		pod.Terminal.Stop()
	}
}

// --- Test OnOutput error handling ---

func TestPodHandlerOnOutputErrorPath(t *testing.T) {
	tempDir := t.TempDir()
	ws, err := workspace.NewManager(tempDir, "")
	if err != nil {
		t.Skipf("Could not create workspace manager: %v", err)
	}

	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 10,
			WorkspaceRoot:         tempDir,
		},
		workspace: ws,
	}
	termManager := terminal.NewManager("/bin/sh", tempDir)
	eventSender := newErrorEventSender(errors.New("send output failed"))
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodStartPayload{
		PodKey:    "error-output-pod",
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

	err = handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Clean up
	pod, ok := store.Get("error-output-pod")
	if ok && pod.Terminal != nil {
		pod.Terminal.Stop()
	}
}

// --- Test handlePodStop with worktree cleanup error ---

func TestPodHandlerStopWithWorktreeCleanupError(t *testing.T) {
	tempDir := t.TempDir()
	ws, err := workspace.NewManager(tempDir, "")
	if err != nil {
		t.Skipf("Could not create workspace manager: %v", err)
	}

	runner := &Runner{
		cfg: &config.Config{
			WorkspaceRoot: tempDir,
		},
		workspace: ws,
	}
	termManager := terminal.NewManager("/bin/sh", tempDir)
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	// Add pod with non-existent worktree path
	store.Put("cleanup-error-pod", &Pod{
		ID:           "cleanup-error-pod",
		PodKey:   "cleanup-error-pod",
		WorktreePath: "/nonexistent/worktree/path",
		Terminal:     nil,
	})

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodStopPayload{PodKey: "cleanup-error-pod"}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodStop,
		Payload: payloadBytes,
	}

	err = handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if store.Count() != 0 {
		t.Errorf("pod count = %d, want 0", store.Count())
	}
}

// --- Test handlePodStart capacity reached ---

func TestPodHandlerStartMaxCapacity(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 1,
			WorkspaceRoot:         "/tmp",
		},
	}
	termManager := terminal.NewManager("/bin/sh", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	// Add pod to reach capacity
	store.Put("existing-pod", &Pod{ID: "existing-pod"})

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodStartPayload{
		PodKey:    "overflow-pod",
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

	// Verify error status was sent
	eventSender.mu.Lock()
	hasErrorStatus := false
	for _, s := range eventSender.statuses {
		if s.status == "error" {
			hasErrorStatus = true
		}
	}
	eventSender.mu.Unlock()

	if !hasErrorStatus {
		t.Error("should have sent error status")
	}
}

// --- Test handlePodList ---

func TestPodHandlerListEmpty(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/sh", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodListPayload{RequestID: "empty-list-req"}
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
	statusCount := len(eventSender.statuses)
	eventSender.mu.Unlock()

	if statusCount != 1 {
		t.Errorf("status count = %d, want 1", statusCount)
	}
}

func TestPodHandlerListMultiplePods(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/sh", "/tmp")
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
		WorktreePath:  "/some/path",
		RepositoryURL: "https://github.com/test/repo",
	})

	handler := NewPodHandler(runner, termManager, eventSender, store)

	payload := PodListPayload{RequestID: "multi-list-req"}
	payloadBytes, _ := json.Marshal(payload)

	msg := &client.Message{
		Type:    client.MessageTypePodList,
		Payload: payloadBytes,
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- Test handlePodExit ---

func TestPodHandlerExitRemovesPod(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/sh", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	// Add pod
	store.Put("exit-pod", &Pod{
		ID:     "exit-pod",
		Status: PodStatusRunning,
	})

	handler := NewPodHandler(runner, termManager, eventSender, store)

	// Trigger exit handler
	handler.handlePodExit("exit-pod", 0)

	// Verify pod was removed
	_, exists := store.Get("exit-pod")
	if exists {
		t.Error("pod should be removed after exit")
	}

	// Verify exit status was sent
	eventSender.mu.Lock()
	hasExitStatus := false
	for _, s := range eventSender.statuses {
		if s.status == "exited" {
			hasExitStatus = true
		}
	}
	eventSender.mu.Unlock()

	if !hasExitStatus {
		t.Error("should have sent exited status")
	}
}

func TestPodHandlerExitNonExistentPod(t *testing.T) {
	runner := &Runner{}
	termManager := terminal.NewManager("/bin/sh", "/tmp")
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	// Should not panic for nonexistent pod
	handler.handlePodExit("nonexistent-pod", 1)
}

// --- Test concurrent pod operations ---

func TestPodHandlerConcurrentPods(t *testing.T) {
	tempDir := t.TempDir()
	ws, err := workspace.NewManager(tempDir, "")
	if err != nil {
		t.Skipf("Could not create workspace manager: %v", err)
	}

	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 100,
			WorkspaceRoot:         tempDir,
		},
		workspace: ws,
	}
	termManager := terminal.NewManager("/bin/sh", tempDir)
	eventSender := newMockEventSender()
	store := NewInMemoryPodStore()

	handler := NewPodHandler(runner, termManager, eventSender, store)

	var wg sync.WaitGroup
	numPods := 5

	// Start multiple pods concurrently
	for i := 0; i < numPods; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			payload := PodStartPayload{
				PodKey:    "concurrent-" + string(rune('A'+idx)),
				AgentType:     "claude-code",
				LaunchCommand: "sleep",
				LaunchArgs:    []string{"0.1"},
				Rows:          24,
				Cols:          80,
			}
			payloadBytes, _ := json.Marshal(payload)

			msg := &client.Message{
				Type:    client.MessageTypePodStart,
				Payload: payloadBytes,
			}

			handler.HandleMessage(context.Background(), msg)
		}(i)
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	t.Logf("Created %d concurrent pods", store.Count())

	// Clean up all pods
	pods := store.All()
	for _, pod := range pods {
		if pod.Terminal != nil {
			pod.Terminal.Stop()
		}
	}

	time.Sleep(200 * time.Millisecond)
}
