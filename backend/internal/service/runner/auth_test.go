package runner

import (
	"context"
	"testing"
)

// --- Authentication Tests ---

func TestAuthenticateRunner(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	// Create token and register runner
	plain, _ := service.CreateRegistrationToken(ctx, 1, 1, "Test Token", nil, nil)
	r, authToken, _ := service.RegisterRunner(ctx, plain, "test-runner", "Test", 5)

	// Authenticate
	authenticated, err := service.AuthenticateRunner(ctx, r.ID, authToken)
	if err != nil {
		t.Fatalf("failed to authenticate runner: %v", err)
	}
	if authenticated == nil {
		t.Fatal("expected non-nil authenticated runner")
	}
	if authenticated.ID != r.ID {
		t.Errorf("expected runner ID %d, got %d", r.ID, authenticated.ID)
	}
}

func TestAuthenticateRunnerInvalidToken(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	// Create token and register runner
	plain, _ := service.CreateRegistrationToken(ctx, 1, 1, "Test Token", nil, nil)
	r, _, _ := service.RegisterRunner(ctx, plain, "test-runner", "Test", 5)

	// Try with invalid token
	_, err := service.AuthenticateRunner(ctx, r.ID, "invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestRegenerateAuthToken(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	plain, _ := service.CreateRegistrationToken(ctx, 1, 1, "Test Token", nil, nil)
	r, oldAuth, _ := service.RegisterRunner(ctx, plain, "test-runner", "Test", 5)

	newAuth, err := service.RegenerateAuthToken(ctx, r.ID)
	if err != nil {
		t.Fatalf("failed to regenerate auth token: %v", err)
	}

	if newAuth == "" {
		t.Fatal("expected non-empty new auth token")
	}
	if newAuth == oldAuth {
		t.Error("new token should be different from old token")
	}

	// Old token should not work
	_, err = service.AuthenticateRunner(ctx, r.ID, oldAuth)
	if err != ErrInvalidToken {
		t.Errorf("expected old token to fail, got %v", err)
	}

	// New token should work
	_, err = service.AuthenticateRunner(ctx, r.ID, newAuth)
	if err != nil {
		t.Errorf("expected new token to work, got %v", err)
	}
}

func TestRegenerateAuthTokenNotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	_, err := service.RegenerateAuthToken(ctx, 99999)
	if err != ErrRunnerNotFound {
		t.Errorf("expected ErrRunnerNotFound, got %v", err)
	}
}

func TestValidateRunnerAuth(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	// Create token and register runner
	plain, _ := service.CreateRegistrationToken(ctx, 1, 1, "Test Token", nil, nil)
	r, authToken, _ := service.RegisterRunner(ctx, plain, "test-runner", "Test", 5)

	// Validate with correct credentials
	validated, err := service.ValidateRunnerAuth(ctx, r.NodeID, authToken)
	if err != nil {
		t.Fatalf("failed to validate runner auth: %v", err)
	}
	if validated == nil {
		t.Fatal("expected non-nil validated runner")
	}
	if validated.ID != r.ID {
		t.Errorf("expected runner ID %d, got %d", r.ID, validated.ID)
	}
}

func TestValidateRunnerAuthNotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	// Try with non-existent node_id
	_, err := service.ValidateRunnerAuth(ctx, "non-existent-runner", "some-token")
	if err != ErrRunnerNotFound {
		t.Errorf("expected ErrRunnerNotFound, got %v", err)
	}
}

func TestValidateRunnerAuthInvalidToken(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	// Create token and register runner
	plain, _ := service.CreateRegistrationToken(ctx, 1, 1, "Test Token", nil, nil)
	r, _, _ := service.RegisterRunner(ctx, plain, "test-runner", "Test", 5)

	// Validate with wrong token
	_, err := service.ValidateRunnerAuth(ctx, r.NodeID, "wrong-token")
	if err != ErrInvalidAuth {
		t.Errorf("expected ErrInvalidAuth, got %v", err)
	}
}

func TestValidateRunnerAuthDisabled(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	// Create token and register runner
	plain, _ := service.CreateRegistrationToken(ctx, 1, 1, "Test Token", nil, nil)
	r, authToken, _ := service.RegisterRunner(ctx, plain, "test-runner", "Test", 5)

	// Disable the runner
	isEnabled := false
	service.UpdateRunner(ctx, r.ID, RunnerUpdateInput{IsEnabled: &isEnabled})

	// Validate should fail for disabled runner
	_, err := service.ValidateRunnerAuth(ctx, r.NodeID, authToken)
	if err == nil {
		t.Error("expected error for disabled runner, got nil")
	}
}
