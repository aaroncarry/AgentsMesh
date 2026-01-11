package runner

import (
	"context"
	"testing"

	"github.com/anthropics/agentmesh/backend/internal/domain/runner"
)

// --- Runner Registration Tests ---

func TestRegisterRunner(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	// Create a registration token first
	plain, err := service.CreateRegistrationToken(ctx, 1, 1, "Test Token", nil, nil)
	if err != nil {
		t.Fatalf("failed to create registration token: %v", err)
	}

	// Register a runner
	r, authToken, err := service.RegisterRunner(ctx, plain, "test-runner-1", "Test Runner", 5)
	if err != nil {
		t.Fatalf("failed to register runner: %v", err)
	}

	if r == nil {
		t.Fatal("expected non-nil runner")
	}
	if authToken == "" {
		t.Fatal("expected non-empty auth token")
	}
	if r.NodeID != "test-runner-1" {
		t.Errorf("expected NodeID 'test-runner-1', got %s", r.NodeID)
	}
	if r.OrganizationID != 1 {
		t.Errorf("expected OrganizationID 1, got %d", r.OrganizationID)
	}
	if r.Status != runner.RunnerStatusOffline {
		t.Errorf("expected Status '%s', got %s", runner.RunnerStatusOffline, r.Status)
	}
	if r.MaxConcurrentPods != 5 {
		t.Errorf("expected MaxConcurrentPods 5, got %d", r.MaxConcurrentPods)
	}

	// Check that token usage count was incremented
	var updatedToken runner.RegistrationToken
	db.First(&updatedToken)
	if updatedToken.UsedCount != 1 {
		t.Errorf("expected UsedCount 1, got %d", updatedToken.UsedCount)
	}
}

func TestRegisterRunnerInvalidToken(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	_, _, err := service.RegisterRunner(ctx, "invalid-token", "test-runner", "Test", 5)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestDeleteRunner(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	// Create runner
	plain, _ := service.CreateRegistrationToken(ctx, 1, 1, "Test Token", nil, nil)
	r, _, _ := service.RegisterRunner(ctx, plain, "test-runner", "Test", 5)

	// Delete runner
	err := service.DeleteRunner(ctx, r.ID)
	if err != nil {
		t.Fatalf("failed to delete runner: %v", err)
	}

	// Verify deletion
	_, err = service.GetRunner(ctx, r.ID)
	if err != ErrRunnerNotFound {
		t.Errorf("expected ErrRunnerNotFound, got %v", err)
	}
}

func TestListRunners(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	// Create multiple runners
	plain, _ := service.CreateRegistrationToken(ctx, 1, 1, "Test Token", nil, nil)
	service.RegisterRunner(ctx, plain, "runner-1", "Runner 1", 5)
	service.RegisterRunner(ctx, plain, "runner-2", "Runner 2", 5)
	service.RegisterRunner(ctx, plain, "runner-3", "Runner 3", 5)

	// List all runners
	runners, err := service.ListRunners(ctx, 1)
	if err != nil {
		t.Fatalf("failed to list runners: %v", err)
	}
	if len(runners) != 3 {
		t.Errorf("expected 3 runners, got %d", len(runners))
	}
}

func TestListAvailableRunners(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	plain, _ := service.CreateRegistrationToken(ctx, 1, 1, "Test Token", nil, nil)

	// Register multiple runners
	r1, _, _ := service.RegisterRunner(ctx, plain, "runner-1", "Runner 1", 5)
	r2, _, _ := service.RegisterRunner(ctx, plain, "runner-2", "Runner 2", 5)

	// Set one runner online
	service.Heartbeat(ctx, r1.ID, 0)

	// List available runners
	runners, err := service.ListAvailableRunners(ctx, 1)
	if err != nil {
		t.Fatalf("failed to list available runners: %v", err)
	}

	if len(runners) != 1 {
		t.Errorf("expected 1 available runner, got %d", len(runners))
	}
	_ = r2
}

func TestSelectAvailableRunner(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	plain, _ := service.CreateRegistrationToken(ctx, 1, 1, "Test Token", nil, nil)
	r1, _, _ := service.RegisterRunner(ctx, plain, "runner-1", "Runner 1", 5)
	r2, _, _ := service.RegisterRunner(ctx, plain, "runner-2", "Runner 2", 5)

	// Make both online
	service.Heartbeat(ctx, r1.ID, 3)
	service.Heartbeat(ctx, r2.ID, 1)

	// Should select r2 (least pods)
	selected, err := service.SelectAvailableRunner(ctx, 1)
	if err != nil {
		t.Fatalf("failed to select available runner: %v", err)
	}
	if selected.ID != r2.ID {
		t.Errorf("expected runner with least pods (r2), got ID %d", selected.ID)
	}
}

func TestSelectAvailableRunnerNoneAvailable(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	// No runners at all
	_, err := service.SelectAvailableRunner(ctx, 1)
	if err != ErrRunnerOffline {
		t.Errorf("expected ErrRunnerOffline, got %v", err)
	}
}
