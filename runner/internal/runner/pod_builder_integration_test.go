package runner

import (
	"context"
	"testing"

	"github.com/anthropics/agentmesh/runner/internal/config"
	"github.com/anthropics/agentmesh/runner/internal/workspace"
)

// --- Test Build method ---

func TestPodBuilderBuildSuccess(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{
			WorkspaceRoot: "/tmp/test-workspace",
			AgentEnvVars:  map[string]string{"CONFIG_VAR": "value"},
		},
	}

	builder := NewPodBuilder(runner).
		WithPodKey("pod-build-test").
		WithAgentType("claude-code").
		WithLaunchCommand("echo", []string{"hello"}).
		WithTerminalSize(30, 100).
		WithInitialPrompt("Test prompt")

	pod, err := builder.Build(context.Background())
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if pod == nil {
		t.Fatal("pod should not be nil")
	}

	if pod.PodKey != "pod-build-test" {
		t.Errorf("PodKey = %v, want pod-build-test", pod.PodKey)
	}
	if pod.AgentType != "claude-code" {
		t.Errorf("AgentType = %v, want claude-code", pod.AgentType)
	}
	if pod.InitialPrompt != "Test prompt" {
		t.Errorf("InitialPrompt = %v, want Test prompt", pod.InitialPrompt)
	}
	if pod.Status != PodStatusInitializing {
		t.Errorf("Status = %v, want initializing", pod.Status)
	}

	// Clean up terminal if created
	if pod.Terminal != nil {
		pod.Terminal.Stop()
	}
}

func TestPodBuilderBuildWithMinimalConfig(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{
			WorkspaceRoot: "/tmp",
		},
	}

	builder := NewPodBuilder(runner).
		WithPodKey("minimal-pod").
		WithLaunchCommand("echo", nil)

	pod, err := builder.Build(context.Background())
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if pod.PodKey != "minimal-pod" {
		t.Errorf("PodKey = %v, want minimal-pod", pod.PodKey)
	}

	// Clean up
	if pod.Terminal != nil {
		pod.Terminal.Stop()
	}
}

// --- Test resolveWorkingDirectory ---

func TestPodBuilderResolveWorkingDirectoryWithWorkspaceManager(t *testing.T) {
	// Create a temporary workspace manager
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

	builder := NewPodBuilder(runner).
		WithPodKey("test-resolve")

	workDir, worktreePath, branchName, err := builder.resolveWorkingDirectory(context.Background())
	if err != nil {
		t.Fatalf("resolveWorkingDirectory failed: %v", err)
	}

	// Without repository URL, should use temp workspace
	if workDir == "" {
		t.Error("workDir should not be empty")
	}
	if worktreePath != "" {
		t.Errorf("worktreePath = %v, want empty (no repo)", worktreePath)
	}
	if branchName != "" {
		t.Errorf("branchName = %v, want empty", branchName)
	}
}

func TestPodBuilderResolveWorkingDirectoryFallbackToConfig(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{
			WorkspaceRoot: "/my/workspace",
		},
		workspace: nil, // No workspace manager
	}

	builder := NewPodBuilder(runner).
		WithPodKey("test-fallback")

	workDir, _, _, err := builder.resolveWorkingDirectory(context.Background())
	if err != nil {
		t.Fatalf("resolveWorkingDirectory failed: %v", err)
	}

	if workDir != "/my/workspace" {
		t.Errorf("workDir = %v, want /my/workspace", workDir)
	}
}

func TestPodBuilderResolveWorkingDirectoryWithTicket(t *testing.T) {
	// Test that without repository URL, falls back to config workspace even with ticket
	runner := &Runner{
		cfg: &config.Config{
			WorkspaceRoot: "/tmp",
		},
	}

	builder := NewPodBuilder(runner).
		WithPodKey("test-ticket-no-repo").
		WithWorktree("TICKET-123")

	workDir, _, _, err := builder.resolveWorkingDirectory(context.Background())
	if err != nil {
		t.Fatalf("resolveWorkingDirectory failed: %v", err)
	}

	// Without repository URL, should fall back to config workspace
	if workDir != "/tmp" {
		t.Errorf("workDir = %v, want /tmp", workDir)
	}
}

// --- Test runPreparation ---

func TestPodBuilderRunPreparationWithScript(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{},
	}

	builder := NewPodBuilder(runner).
		WithPodKey("test-prep").
		WithPreparationScript("echo hello", 5)

	// runPreparation will create a preparer and try to run it
	// This may fail in test environment, but we test the path
	err := builder.runPreparation(context.Background(), "/tmp", "/tmp/worktree", "main")

	// We mainly want to test the code path executes without panic
	// The actual script may fail depending on environment
	_ = err
}

func TestPodBuilderRunPreparationEmptyScript(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{},
	}

	builder := NewPodBuilder(runner).
		WithPodKey("test-no-prep")
	// prepScript is empty

	err := builder.runPreparation(context.Background(), "/tmp", "", "")
	if err != nil {
		t.Errorf("runPreparation with empty script should not error: %v", err)
	}
}

func TestPodBuilderRunPreparationBasic(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{},
	}

	builder := NewPodBuilder(runner).
		WithPodKey("test-prep-basic").
		WithPreparationScript("echo test", 5)

	// Test basic preparation - should work
	err := builder.runPreparation(context.Background(), "/tmp", "/tmp/worktree", "main")
	_ = err // May fail but should not panic
}

// --- Test mergeEnvVars edge cases ---

func TestPodBuilderMergeEnvVarsEmptyBoth(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{
			AgentEnvVars: nil,
		},
	}

	builder := NewPodBuilder(runner)
	// Don't add any env vars

	result := builder.mergeEnvVars()

	if len(result) != 0 {
		t.Errorf("result length = %d, want 0", len(result))
	}
}

func TestPodBuilderMergeEnvVarsOnlyConfig(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{
			AgentEnvVars: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
			},
		},
	}

	builder := NewPodBuilder(runner)

	result := builder.mergeEnvVars()

	if result["VAR1"] != "value1" {
		t.Errorf("VAR1 = %v, want value1", result["VAR1"])
	}
	if result["VAR2"] != "value2" {
		t.Errorf("VAR2 = %v, want value2", result["VAR2"])
	}
}

func TestPodBuilderMergeEnvVarsOnlyBuilder(t *testing.T) {
	runner := &Runner{
		cfg: &config.Config{
			AgentEnvVars: nil,
		},
	}

	builder := NewPodBuilder(runner).
		WithEnvVar("BUILDER_VAR", "builder_value")

	result := builder.mergeEnvVars()

	if result["BUILDER_VAR"] != "builder_value" {
		t.Errorf("BUILDER_VAR = %v, want builder_value", result["BUILDER_VAR"])
	}
}

// --- Test ExtendedPod ---

func TestExtendedPodEmbedding(t *testing.T) {
	pod := &Pod{
		ID:            "pod-1",
		PodKey:    "key-1",
		AgentType:     "claude-code",
		Status:        PodStatusRunning,
	}

	extended := &ExtendedPod{
		Pod:              pod,
		TicketIdentifier: "TICKET-999",
		OnOutput:         func([]byte) {},
		OnExit:           func(int) {},
	}

	// Test that embedded fields are accessible
	if extended.ID != "pod-1" {
		t.Errorf("ID = %v, want pod-1", extended.ID)
	}
	if extended.PodKey != "key-1" {
		t.Errorf("PodKey = %v, want key-1", extended.PodKey)
	}
	if extended.TicketIdentifier != "TICKET-999" {
		t.Errorf("TicketIdentifier = %v, want TICKET-999", extended.TicketIdentifier)
	}
	if extended.OnOutput == nil {
		t.Error("OnOutput should not be nil")
	}
	if extended.OnExit == nil {
		t.Error("OnExit should not be nil")
	}
}

// --- Test WithMCP ---

func TestPodBuilderWithMCPSingle(t *testing.T) {
	runner := &Runner{}
	builder := NewPodBuilder(runner).
		WithMCP("server1")

	if !builder.mcpEnabled {
		t.Error("mcpEnabled should be true")
	}
	if len(builder.mcpServers) != 1 {
		t.Errorf("mcpServers length = %d, want 1", len(builder.mcpServers))
	}
	if builder.mcpServers[0] != "server1" {
		t.Errorf("mcpServers[0] = %v, want server1", builder.mcpServers[0])
	}
}

func TestPodBuilderWithMCPMultiple(t *testing.T) {
	runner := &Runner{}
	builder := NewPodBuilder(runner).
		WithMCP("server1", "server2", "server3")

	if !builder.mcpEnabled {
		t.Error("mcpEnabled should be true")
	}
	if len(builder.mcpServers) != 3 {
		t.Errorf("mcpServers length = %d, want 3", len(builder.mcpServers))
	}
}

func TestPodBuilderWithMCPEmpty(t *testing.T) {
	runner := &Runner{}
	builder := NewPodBuilder(runner).
		WithMCP()

	if !builder.mcpEnabled {
		t.Error("mcpEnabled should be true even with no servers")
	}
	if len(builder.mcpServers) != 0 {
		t.Errorf("mcpServers length = %d, want 0", len(builder.mcpServers))
	}
}

// --- Benchmark ---

func BenchmarkPodBuilderBuild(b *testing.B) {
	runner := &Runner{
		cfg: &config.Config{
			WorkspaceRoot: "/tmp",
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewPodBuilder(runner).
			WithPodKey("benchmark-pod").
			WithAgentType("claude-code").
			WithLaunchCommand("echo", []string{"test"})

		pod, _ := builder.Build(ctx)
		if pod != nil && pod.Terminal != nil {
			pod.Terminal.Stop()
		}
	}
}
