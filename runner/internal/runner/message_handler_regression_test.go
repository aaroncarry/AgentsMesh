package runner

import (
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/anthropics/agentsmesh/runner/internal/config"
	"github.com/anthropics/agentsmesh/runner/internal/terminal/aggregator"
)

// Regression tests for issues found during deep review rounds 4-5.
// Each test targets a specific bug fix to prevent future regressions.

// --- OnTerminatePod with nil aggregator must not panic ---

func TestOnTerminatePod_NilAggregator(t *testing.T) {
	store := NewInMemoryPodStore()
	mockConn := client.NewMockConnection()
	runner := &Runner{cfg: &config.Config{}}
	handler := NewRunnerMessageHandler(runner, store, mockConn)

	// Pod without aggregator should not panic.
	store.Put("no-agg-pod", &Pod{PodKey: "no-agg-pod"})

	err := handler.OnTerminatePod(client.TerminatePodRequest{PodKey: "no-agg-pod"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- M1: OnTerminalResize must check pod.Terminal != nil ---
// Bug: OnTerminalResize would panic on nil pointer dereference when
// pod.Terminal was nil (before initialization or after teardown).

func TestOnTerminalResize_NilTerminal(t *testing.T) {
	store := NewInMemoryPodStore()
	mockConn := client.NewMockConnection()
	runner := &Runner{cfg: &config.Config{}}
	handler := NewRunnerMessageHandler(runner, store, mockConn)

	// Pod with nil Terminal (e.g., still initializing).
	store.Put("resize-nil-pod", &Pod{
		PodKey:   "resize-nil-pod",
		Terminal: nil,
	})

	err := handler.OnTerminalResize(client.TerminalResizeRequest{
		PodKey: "resize-nil-pod",
		Cols:   120,
		Rows:   40,
	})

	if err == nil {
		t.Fatal("expected error for nil terminal")
	}
	if !contains(err.Error(), "terminal not initialized") {
		t.Errorf("error = %v, want containing 'terminal not initialized'", err)
	}
}

// --- M1b: OnTerminalRedraw must check pod.Terminal != nil ---
// Bug: OnTerminalRedraw would panic on nil pointer dereference when
// pod.Terminal was nil (same class of bug as OnTerminalResize).

func TestOnTerminalRedraw_NilTerminal(t *testing.T) {
	store := NewInMemoryPodStore()
	mockConn := client.NewMockConnection()
	runner := &Runner{cfg: &config.Config{}}
	handler := NewRunnerMessageHandler(runner, store, mockConn)

	store.Put("redraw-nil-pod", &Pod{
		PodKey:   "redraw-nil-pod",
		Terminal: nil,
	})

	err := handler.OnTerminalRedraw(client.TerminalRedrawRequest{
		PodKey: "redraw-nil-pod",
	})

	if err == nil {
		t.Fatal("expected error for nil terminal on redraw")
	}
	if !contains(err.Error(), "terminal not initialized") {
		t.Errorf("error = %v, want containing 'terminal not initialized'", err)
	}
}

// --- Cleanup path consistency: both paths must produce pod_terminated event ---

func TestTerminationPaths_BothSendEvent(t *testing.T) {
	// Verify both exit paths (natural exit and server-initiated) send
	// the pod_terminated event so the backend always gets notified.

	t.Run("natural_exit", func(t *testing.T) {
		store := NewInMemoryPodStore()
		mockConn := client.NewMockConnection()
		runner := &Runner{cfg: &config.Config{}}
		handler := NewRunnerMessageHandler(runner, store, mockConn)

		store.Put("exit-pod", &Pod{PodKey: "exit-pod", Status: PodStatusRunning})
		handler.createExitHandler("exit-pod")(0)

		assertHasEvent(t, mockConn, client.MsgTypePodTerminated)
	})

	t.Run("server_terminate", func(t *testing.T) {
		store := NewInMemoryPodStore()
		mockConn := client.NewMockConnection()
		runner := &Runner{cfg: &config.Config{}}
		handler := NewRunnerMessageHandler(runner, store, mockConn)

		store.Put("term-pod", &Pod{PodKey: "term-pod", Status: PodStatusRunning})
		err := handler.OnTerminatePod(client.TerminatePodRequest{PodKey: "term-pod"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assertHasEvent(t, mockConn, client.MsgTypePodTerminated)
	})
}

// --- Atomic podStore.Delete prevents double cleanup ---

func TestConcurrentExitAndTerminate_OnlyOneCleanup(t *testing.T) {
	store := NewInMemoryPodStore()
	mockConn := client.NewMockConnection()
	runner := &Runner{cfg: &config.Config{}}
	handler := NewRunnerMessageHandler(runner, store, mockConn)

	agg := aggregator.NewSmartAggregator(nil)
	store.Put("race-pod", &Pod{
		PodKey:     "race-pod",
		Status:     PodStatusRunning,
		Aggregator: agg,
	})

	// Simulate exit handler winning the race.
	exitHandler := handler.createExitHandler("race-pod")
	exitHandler(0)

	// Server-initiated terminate arrives after — pod already removed.
	err := handler.OnTerminatePod(client.TerminatePodRequest{PodKey: "race-pod"})
	if err == nil {
		t.Error("expected error when pod already removed")
	}
	if !contains(err.Error(), "pod not found") {
		t.Errorf("error = %v, want 'pod not found'", err)
	}
}

// --- Helper ---

func assertHasEvent(t *testing.T, mockConn *client.MockConnection, eventType client.MessageType) {
	t.Helper()
	events := mockConn.GetEvents()
	for _, e := range events {
		if e.Type == eventType {
			return
		}
	}
	t.Errorf("expected event %q not found in %d events", eventType, len(events))
}
