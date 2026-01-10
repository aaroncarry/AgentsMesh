package runner

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/anthropics/agentmesh/backend/internal/domain/runner"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrRunnerNotFound      = errors.New("runner not found")
	ErrRunnerOffline       = errors.New("runner is offline")
	ErrInvalidToken        = errors.New("invalid registration token")
	ErrInvalidAuth         = errors.New("invalid runner authentication")
	ErrTokenExpired        = errors.New("registration token expired")
	ErrTokenExhausted      = errors.New("registration token usage exhausted")
	ErrRunnerAlreadyExists = errors.New("runner already exists")
)

// Service handles runner operations
type Service struct {
	db            *gorm.DB
	activeRunners sync.Map // map[runnerID]*ActiveRunner
}

// ActiveRunner represents an active runner connection
type ActiveRunner struct {
	Runner   *runner.Runner
	LastPing time.Time
	PodCount int
}

// NewService creates a new runner service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// CreateRegistrationToken creates a new registration token
func (s *Service) CreateRegistrationToken(ctx context.Context, orgID, userID int64, description string, maxUses *int, expiresAt *time.Time) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	// Hash the token for storage
	tokenHash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	regToken := &runner.RegistrationToken{
		OrganizationID: orgID,
		TokenHash:      string(tokenHash),
		Description:    description,
		CreatedByID:    userID,
		IsActive:       true,
		MaxUses:        maxUses,
		UsedCount:      0,
		ExpiresAt:      expiresAt,
	}

	if err := s.db.WithContext(ctx).Create(regToken).Error; err != nil {
		return "", err
	}

	return token, nil
}

// ValidateRegistrationToken validates a registration token
func (s *Service) ValidateRegistrationToken(ctx context.Context, token string) (*runner.RegistrationToken, error) {
	var tokens []runner.RegistrationToken
	if err := s.db.WithContext(ctx).Where("is_active = ?", true).Find(&tokens).Error; err != nil {
		return nil, err
	}

	for _, t := range tokens {
		if err := bcrypt.CompareHashAndPassword([]byte(t.TokenHash), []byte(token)); err == nil {
			// Check expiration
			if t.ExpiresAt != nil && t.ExpiresAt.Before(time.Now()) {
				return nil, ErrTokenExpired
			}

			// Check usage
			if t.MaxUses != nil && t.UsedCount >= *t.MaxUses {
				return nil, ErrTokenExhausted
			}

			return &t, nil
		}
	}

	return nil, ErrInvalidToken
}

// RegisterRunner registers a new runner
func (s *Service) RegisterRunner(ctx context.Context, token, nodeID, description string, maxPods int) (*runner.Runner, string, error) {
	// Validate token
	regToken, err := s.ValidateRegistrationToken(ctx, token)
	if err != nil {
		return nil, "", err
	}

	// Check if runner already exists
	var existing runner.Runner
	if err := s.db.WithContext(ctx).Where("organization_id = ? AND node_id = ?", regToken.OrganizationID, nodeID).First(&existing).Error; err == nil {
		return nil, "", ErrRunnerAlreadyExists
	}

	// Generate auth token
	authTokenBytes := make([]byte, 32)
	if _, err := rand.Read(authTokenBytes); err != nil {
		return nil, "", err
	}
	authToken := hex.EncodeToString(authTokenBytes)

	authTokenHash, err := bcrypt.GenerateFromPassword([]byte(authToken), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	// Create runner
	r := &runner.Runner{
		OrganizationID:        regToken.OrganizationID,
		NodeID:                nodeID,
		Description:           description,
		AuthTokenHash:         string(authTokenHash),
		Status:                runner.RunnerStatusOffline,
		MaxConcurrentPods: maxPods,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(r).Error; err != nil {
			return err
		}

		// Increment token usage
		return tx.Model(regToken).Update("used_count", gorm.Expr("used_count + 1")).Error
	})

	if err != nil {
		return nil, "", err
	}

	return r, authToken, nil
}

// AuthenticateRunner authenticates a runner by its auth token
func (s *Service) AuthenticateRunner(ctx context.Context, runnerID int64, authToken string) (*runner.Runner, error) {
	var r runner.Runner
	if err := s.db.WithContext(ctx).First(&r, runnerID).Error; err != nil {
		return nil, ErrRunnerNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(r.AuthTokenHash), []byte(authToken)); err != nil {
		return nil, ErrInvalidToken
	}

	return &r, nil
}

// UpdateRunnerStatus updates runner status
func (s *Service) UpdateRunnerStatus(ctx context.Context, runnerID int64, status string) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&runner.Runner{}).Where("id = ?", runnerID).Updates(map[string]interface{}{
		"status":         status,
		"last_heartbeat": now,
	}).Error
}

// Heartbeat updates runner heartbeat
func (s *Service) Heartbeat(ctx context.Context, runnerID int64, currentPods int) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&runner.Runner{}).Where("id = ?", runnerID).Updates(map[string]interface{}{
		"last_heartbeat":   now,
		"current_pods": currentPods,
		"status":           runner.RunnerStatusOnline,
	}).Error
}

// GetRunner returns a runner by ID
func (s *Service) GetRunner(ctx context.Context, runnerID int64) (*runner.Runner, error) {
	var r runner.Runner
	if err := s.db.WithContext(ctx).First(&r, runnerID).Error; err != nil {
		return nil, ErrRunnerNotFound
	}
	return &r, nil
}

// ListRunners returns runners for an organization
func (s *Service) ListRunners(ctx context.Context, orgID int64) ([]*runner.Runner, error) {
	var runners []*runner.Runner
	if err := s.db.WithContext(ctx).Where("organization_id = ?", orgID).Find(&runners).Error; err != nil {
		return nil, err
	}
	return runners, nil
}

// ListAvailableRunners returns online runners that can accept pods
func (s *Service) ListAvailableRunners(ctx context.Context, orgID int64) ([]*runner.Runner, error) {
	var runners []*runner.Runner
	if err := s.db.WithContext(ctx).
		Where("organization_id = ? AND status = ? AND is_enabled = ? AND current_pods < max_concurrent_pods", orgID, runner.RunnerStatusOnline, true).
		Find(&runners).Error; err != nil {
		return nil, err
	}
	return runners, nil
}

// DeleteRunner deletes a runner
func (s *Service) DeleteRunner(ctx context.Context, runnerID int64) error {
	return s.db.WithContext(ctx).Delete(&runner.Runner{}, runnerID).Error
}

// RunnerUpdateInput represents input for updating a runner
type RunnerUpdateInput struct {
	Description       *string `json:"description"`
	MaxConcurrentPods *int    `json:"max_concurrent_pods"`
	IsEnabled         *bool   `json:"is_enabled"`
}

// UpdateRunner updates a runner's configuration
func (s *Service) UpdateRunner(ctx context.Context, runnerID int64, input RunnerUpdateInput) (*runner.Runner, error) {
	var r runner.Runner
	if err := s.db.WithContext(ctx).First(&r, runnerID).Error; err != nil {
		return nil, ErrRunnerNotFound
	}

	updates := make(map[string]interface{})
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.MaxConcurrentPods != nil {
		updates["max_concurrent_pods"] = *input.MaxConcurrentPods
	}
	if input.IsEnabled != nil {
		updates["is_enabled"] = *input.IsEnabled
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&r).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	// Reload the runner
	if err := s.db.WithContext(ctx).First(&r, runnerID).Error; err != nil {
		return nil, err
	}

	return &r, nil
}

// RegenerateAuthToken generates a new authentication token for a runner
func (s *Service) RegenerateAuthToken(ctx context.Context, runnerID int64) (string, error) {
	var r runner.Runner
	if err := s.db.WithContext(ctx).First(&r, runnerID).Error; err != nil {
		return "", ErrRunnerNotFound
	}

	// Generate new auth token
	authTokenBytes := make([]byte, 32)
	if _, err := rand.Read(authTokenBytes); err != nil {
		return "", err
	}
	authToken := hex.EncodeToString(authTokenBytes)

	authTokenHash, err := bcrypt.GenerateFromPassword([]byte(authToken), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Update the runner with the new token hash
	if err := s.db.WithContext(ctx).Model(&r).Update("auth_token_hash", string(authTokenHash)).Error; err != nil {
		return "", err
	}

	return authToken, nil
}

// MarkOfflineRunners marks runners as offline if no heartbeat received
func (s *Service) MarkOfflineRunners(ctx context.Context, timeout time.Duration) error {
	threshold := time.Now().Add(-timeout)
	return s.db.WithContext(ctx).Model(&runner.Runner{}).
		Where("status = ? AND last_heartbeat < ?", runner.RunnerStatusOnline, threshold).
		Update("status", runner.RunnerStatusOffline).Error
}

// ListRegistrationTokens lists registration tokens for an organization
func (s *Service) ListRegistrationTokens(ctx context.Context, orgID int64) ([]*runner.RegistrationToken, error) {
	var tokens []*runner.RegistrationToken
	if err := s.db.WithContext(ctx).Where("organization_id = ?", orgID).Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

// RevokeRegistrationToken revokes a registration token
func (s *Service) RevokeRegistrationToken(ctx context.Context, tokenID int64) error {
	return s.db.WithContext(ctx).Model(&runner.RegistrationToken{}).
		Where("id = ?", tokenID).
		Update("is_active", false).Error
}

// UpdateHeartbeat updates runner heartbeat with authentication
func (s *Service) UpdateHeartbeat(ctx context.Context, runnerID int64, authToken string, currentPods int, version string) error {
	// Verify runner authentication
	r, err := s.AuthenticateRunner(ctx, runnerID, authToken)
	if err != nil {
		if err == ErrInvalidToken {
			return ErrInvalidAuth
		}
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"last_heartbeat":   now,
		"current_pods": currentPods,
		"status":           runner.RunnerStatusOnline,
	}
	if version != "" {
		updates["runner_version"] = version
	}

	return s.db.WithContext(ctx).Model(r).Updates(updates).Error
}

// HeartbeatPodInfo represents a pod reported in heartbeat
// Note: This is a duplicate of HeartbeatPod in connection_manager.go for legacy API compatibility
type HeartbeatPodInfo struct {
	PodKey      string `json:"pod_key"`
	Status      string `json:"status,omitempty"`
	AgentStatus string `json:"agent_status,omitempty"`
}

// UpdateHeartbeatWithPods updates runner heartbeat with pod reconciliation
func (s *Service) UpdateHeartbeatWithPods(ctx context.Context, runnerID int64, pods []HeartbeatPodInfo, version string) error {
	var r runner.Runner
	if err := s.db.WithContext(ctx).First(&r, runnerID).Error; err != nil {
		return ErrRunnerNotFound
	}

	now := time.Now()
	updates := map[string]interface{}{
		"last_heartbeat": now,
		"current_pods":   len(pods),
		"status":         runner.RunnerStatusOnline,
	}
	if version != "" {
		updates["runner_version"] = version
	}

	// Update active runner in memory
	s.activeRunners.Store(runnerID, &ActiveRunner{
		Runner:   &r,
		LastPing: now,
		PodCount: len(pods),
	})

	return s.db.WithContext(ctx).Model(&r).Updates(updates).Error
}

// SelectAvailableRunner selects an available runner using least-pods strategy
func (s *Service) SelectAvailableRunner(ctx context.Context, orgID int64) (*runner.Runner, error) {
	var runners []*runner.Runner
	if err := s.db.WithContext(ctx).
		Where("organization_id = ? AND status = ? AND is_enabled = ? AND current_pods < max_concurrent_pods", orgID, runner.RunnerStatusOnline, true).
		Order("current_pods ASC").
		Find(&runners).Error; err != nil {
		return nil, err
	}

	if len(runners) == 0 {
		return nil, ErrRunnerOffline
	}

	// Return the runner with least pods
	return runners[0], nil
}

// IncrementPods increments the pod count for a runner
func (s *Service) IncrementPods(ctx context.Context, runnerID int64) error {
	return s.db.WithContext(ctx).Exec(
		"UPDATE runners SET current_pods = current_pods + 1 WHERE id = ?",
		runnerID,
	).Error
}

// DecrementPods decrements the pod count for a runner
func (s *Service) DecrementPods(ctx context.Context, runnerID int64) error {
	return s.db.WithContext(ctx).Exec(
		"UPDATE runners SET current_pods = GREATEST(current_pods - 1, 0) WHERE id = ?",
		runnerID,
	).Error
}

// IsConnected checks if a runner has an active connection
func (s *Service) IsConnected(runnerID int64) bool {
	_, exists := s.activeRunners.Load(runnerID)
	return exists
}

// MarkConnected marks a runner as connected
func (s *Service) MarkConnected(ctx context.Context, runnerID int64) error {
	r, err := s.GetRunner(ctx, runnerID)
	if err != nil {
		return err
	}

	now := time.Now()
	s.activeRunners.Store(runnerID, &ActiveRunner{
		Runner:   r,
		LastPing: now,
		PodCount: r.CurrentPods,
	})

	return s.UpdateRunnerStatus(ctx, runnerID, runner.RunnerStatusOnline)
}

// MarkDisconnected marks a runner as disconnected
func (s *Service) MarkDisconnected(ctx context.Context, runnerID int64) error {
	s.activeRunners.Delete(runnerID)
	return s.UpdateRunnerStatus(ctx, runnerID, runner.RunnerStatusOffline)
}

// UpdateHostInfo updates runner host information
func (s *Service) UpdateHostInfo(ctx context.Context, runnerID int64, hostInfo map[string]interface{}) error {
	return s.db.WithContext(ctx).Model(&runner.Runner{}).
		Where("id = ?", runnerID).
		Update("host_info", hostInfo).Error
}

// RunnerUpdateFunc is a callback for runner status updates
type RunnerUpdateFunc func(*runner.Runner)

// SubscribeStatusChanges subscribes to runner status changes and returns an unsubscribe function
func (s *Service) SubscribeStatusChanges(ctx context.Context, callback RunnerUpdateFunc) (func(), error) {
	// In a real implementation, this would use Redis pub/sub or similar
	// For now, return a simple unsubscribe function
	return func() {}, nil
}

// ValidateRunnerAuth validates runner authentication by node_id and auth token
// Returns the runner if authentication is successful
func (s *Service) ValidateRunnerAuth(ctx context.Context, nodeID, authToken string) (*runner.Runner, error) {
	var r runner.Runner
	if err := s.db.WithContext(ctx).Where("node_id = ?", nodeID).First(&r).Error; err != nil {
		return nil, ErrRunnerNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(r.AuthTokenHash), []byte(authToken)); err != nil {
		return nil, ErrInvalidAuth
	}

	if !r.IsEnabled {
		return nil, errors.New("runner is disabled")
	}

	return &r, nil
}

// SetRunnerStatus sets the runner status (alias for UpdateRunnerStatus)
func (s *Service) SetRunnerStatus(ctx context.Context, runnerID int64, status string) error {
	return s.UpdateRunnerStatus(ctx, runnerID, status)
}
