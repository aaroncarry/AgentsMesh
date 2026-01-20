package runner

import (
	"log/slog"
	"testing"
)

func TestTerminalRouter_GetRecentOutput(t *testing.T) {
	logger := slog.Default()
	cm := NewRunnerConnectionManager(logger)
	tr := NewTerminalRouter(cm, logger)

	// Test with non-existent pod
	output := tr.GetRecentOutput("nonexistent-pod", 10)
	if output != nil {
		t.Errorf("Expected nil for nonexistent pod, got %s", string(output))
	}

	// Register a pod and feed data
	tr.RegisterPodWithSize("pod-1", 1, 80, 24)

	// Feed some data
	shard := tr.getShard("pod-1")
	shard.mu.RLock()
	vt := shard.virtualTerminals["pod-1"]
	shard.mu.RUnlock()
	if vt != nil {
		vt.Feed([]byte("Hello World\n"))
	}

	// Get output
	output = tr.GetRecentOutput("pod-1", 10)
	if output == nil {
		t.Error("Expected output, got nil")
	}
}

func TestTerminalRouter_GetScreenSnapshot(t *testing.T) {
	logger := slog.Default()
	cm := NewRunnerConnectionManager(logger)
	tr := NewTerminalRouter(cm, logger)

	// Test with non-existent pod
	snapshot := tr.GetScreenSnapshot("nonexistent-pod")
	if snapshot != "" {
		t.Errorf("Expected empty string for nonexistent pod, got %s", snapshot)
	}

	// Register a pod
	tr.RegisterPodWithSize("pod-1", 1, 80, 24)

	// Get snapshot (empty initially but should not error)
	snapshot = tr.GetScreenSnapshot("pod-1")
	// May be empty for new terminal
}

func TestTerminalRouter_GetCursorPosition(t *testing.T) {
	logger := slog.Default()
	cm := NewRunnerConnectionManager(logger)
	tr := NewTerminalRouter(cm, logger)

	// Test with non-existent pod
	row, col := tr.GetCursorPosition("nonexistent-pod")
	if row != 0 || col != 0 {
		t.Errorf("Expected (0, 0) for nonexistent pod, got (%d, %d)", row, col)
	}

	// Register a pod
	tr.RegisterPodWithSize("pod-1", 1, 80, 24)

	// Get cursor position (should be 0, 0 initially)
	row, col = tr.GetCursorPosition("pod-1")
	if row != 0 || col != 0 {
		t.Errorf("Expected (0, 0), got (%d, %d)", row, col)
	}
}

func TestTerminalRouter_ClearTerminal(t *testing.T) {
	logger := slog.Default()
	cm := NewRunnerConnectionManager(logger)
	tr := NewTerminalRouter(cm, logger)

	// Should not panic for non-existent pod
	tr.ClearTerminal("nonexistent-pod")

	// Register a pod
	tr.RegisterPodWithSize("pod-1", 1, 80, 24)

	// Clear should not panic
	tr.ClearTerminal("pod-1")
}
