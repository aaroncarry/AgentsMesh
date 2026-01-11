package runner

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/anthropics/agentmesh/backend/internal/domain/agentpod"
	runnerDomain "github.com/anthropics/agentmesh/backend/internal/domain/runner"
	"gorm.io/gorm"
)

// PodCoordinator coordinates pod lifecycle events between backend and runners
type PodCoordinator struct {
	db                *gorm.DB
	connectionManager *ConnectionManager
	terminalRouter    *TerminalRouter
	logger            *slog.Logger

	// Callbacks
	onStatusChange func(podKey string, status string, agentStatus string)

	// Cache for capabilities hash to avoid unnecessary DB updates
	capabilitiesHashCache map[int64]string
}

// NewPodCoordinator creates a new pod coordinator
func NewPodCoordinator(
	db *gorm.DB,
	cm *ConnectionManager,
	tr *TerminalRouter,
	logger *slog.Logger,
) *PodCoordinator {
	pc := &PodCoordinator{
		db:                    db,
		connectionManager:     cm,
		terminalRouter:        tr,
		logger:                logger,
		capabilitiesHashCache: make(map[int64]string),
	}

	// Set up callbacks from connection manager
	cm.SetHeartbeatCallback(pc.handleHeartbeat)
	cm.SetPodCreatedCallback(pc.handlePodCreated)
	cm.SetPodTerminatedCallback(pc.handlePodTerminated)
	cm.SetAgentStatusCallback(pc.handleAgentStatus)
	cm.SetDisconnectCallback(pc.handleRunnerDisconnect)

	return pc
}

// SetStatusChangeCallback sets the callback for status changes
func (pc *PodCoordinator) SetStatusChangeCallback(fn func(podKey string, status string, agentStatus string)) {
	pc.onStatusChange = fn
}

// handleHeartbeat handles heartbeat from a runner
func (pc *PodCoordinator) handleHeartbeat(runnerID int64, data *HeartbeatData) {
	ctx := context.Background()

	// Update runner
	now := time.Now()
	updates := map[string]interface{}{
		"last_heartbeat": now,
		"current_pods":   len(data.Pods),
		"status":         "online",
	}
	if data.RunnerVersion != "" {
		updates["runner_version"] = data.RunnerVersion
	}

	// Only update capabilities if they changed (using hash comparison)
	if len(data.Capabilities) > 0 {
		newHash := pc.hashCapabilities(data.Capabilities)
		cachedHash, ok := pc.capabilitiesHashCache[runnerID]
		if !ok || cachedHash != newHash {
			updates["capabilities"] = runnerDomain.Capabilities(data.Capabilities)
			pc.capabilitiesHashCache[runnerID] = newHash
			pc.logger.Debug("updating runner capabilities",
				"runner_id", runnerID,
				"capabilities_count", len(data.Capabilities))
		}
	}

	if err := pc.db.WithContext(ctx).
		Table("runners").
		Where("id = ?", runnerID).
		Updates(updates).Error; err != nil {
		pc.logger.Error("failed to update runner heartbeat",
			"runner_id", runnerID,
			"error", err)
		return
	}

	// Reconcile pods
	reportedPodKeys := make(map[string]bool)
	for _, p := range data.Pods {
		reportedPodKeys[p.PodKey] = true
	}

	pc.reconcilePods(ctx, runnerID, reportedPodKeys)
}

// reconcilePods syncs pod status between runner heartbeat and database
func (pc *PodCoordinator) reconcilePods(ctx context.Context, runnerID int64, reportedPods map[string]bool) {
	now := time.Now()

	// First, ensure all reported pods are registered with terminal router
	// and restore any orphaned pods that runner reports as active
	for podKey := range reportedPods {
		// Always register pod with terminal router (idempotent operation)
		// This ensures routing works even after backend restart
		pc.terminalRouter.RegisterPod(podKey, runnerID)

		// Try to restore if pod is orphaned
		result := pc.db.WithContext(ctx).
			Model(&agentpod.Pod{}).
			Where("pod_key = ? AND runner_id = ? AND status = ?", podKey, runnerID, agentpod.StatusOrphaned).
			Updates(map[string]interface{}{
				"status":        agentpod.StatusRunning,
				"finished_at":   nil,
				"last_activity": now,
			})
		if result.Error != nil {
			pc.logger.Error("failed to restore orphaned pod",
				"pod_key", podKey,
				"error", result.Error)
		} else if result.RowsAffected > 0 {
			pc.logger.Info("restored orphaned pod reported by runner",
				"pod_key", podKey,
				"runner_id", runnerID)
		}
	}

	// Get active pods for this runner from database
	var pods []agentpod.Pod
	if err := pc.db.WithContext(ctx).
		Where("runner_id = ? AND status IN ?", runnerID, []string{agentpod.StatusRunning, agentpod.StatusInitializing}).
		Find(&pods).Error; err != nil {
		pc.logger.Error("failed to get pods for reconciliation",
			"runner_id", runnerID,
			"error", err)
		return
	}

	// Mark pods that are in DB but not reported by runner as orphaned
	for _, p := range pods {
		if !reportedPods[p.PodKey] {
			if err := pc.db.WithContext(ctx).
				Model(&p).
				Updates(map[string]interface{}{
					"status":      agentpod.StatusOrphaned,
					"finished_at": now,
				}).Error; err != nil {
				pc.logger.Error("failed to mark pod as orphaned",
					"pod_key", p.PodKey,
					"error", err)
			} else {
				pc.logger.Warn("pod marked as orphaned (not reported by runner)",
					"pod_key", p.PodKey,
					"runner_id", runnerID)
				// Unregister from terminal router
				pc.terminalRouter.UnregisterPod(p.PodKey)
			}
		}
	}
}

// handlePodCreated handles pod creation event from runner
func (pc *PodCoordinator) handlePodCreated(runnerID int64, data *PodCreatedData) {
	ctx := context.Background()

	now := time.Now()
	updates := map[string]interface{}{
		"pty_pid":       data.Pid,
		"status":        agentpod.StatusRunning,
		"started_at":    now,
		"last_activity": now,
	}

	if data.BranchName != "" {
		updates["branch_name"] = data.BranchName
	}
	if data.WorktreePath != "" {
		updates["worktree_path"] = data.WorktreePath
	}

	if err := pc.db.WithContext(ctx).
		Model(&agentpod.Pod{}).
		Where("pod_key = ?", data.PodKey).
		Updates(updates).Error; err != nil {
		pc.logger.Error("failed to update pod on creation",
			"pod_key", data.PodKey,
			"error", err)
		return
	}

	// Register with terminal router
	pc.terminalRouter.RegisterPod(data.PodKey, runnerID)

	pc.logger.Info("pod created",
		"pod_key", data.PodKey,
		"runner_id", runnerID,
		"pid", data.Pid,
		"branch", data.BranchName)

	// Notify status change
	if pc.onStatusChange != nil {
		pc.onStatusChange(data.PodKey, agentpod.StatusRunning, "")
	}
}

// handlePodTerminated handles pod termination event from runner
func (pc *PodCoordinator) handlePodTerminated(runnerID int64, data *PodTerminatedData) {
	ctx := context.Background()

	now := time.Now()
	if err := pc.db.WithContext(ctx).
		Model(&agentpod.Pod{}).
		Where("pod_key = ?", data.PodKey).
		Updates(map[string]interface{}{
			"status":      agentpod.StatusCompleted,
			"finished_at": now,
			"pty_pid":     nil,
		}).Error; err != nil {
		pc.logger.Error("failed to update pod on termination",
			"pod_key", data.PodKey,
			"error", err)
		return
	}

	// Decrement runner pod count
	pc.db.WithContext(ctx).Exec(
		"UPDATE runners SET current_pods = GREATEST(current_pods - 1, 0) WHERE id = ?",
		runnerID,
	)

	// Unregister from terminal router
	pc.terminalRouter.UnregisterPod(data.PodKey)

	pc.logger.Info("pod terminated",
		"pod_key", data.PodKey,
		"runner_id", runnerID,
		"exit_code", data.ExitCode)

	// Notify status change
	if pc.onStatusChange != nil {
		pc.onStatusChange(data.PodKey, agentpod.StatusCompleted, "")
	}
}

// handleAgentStatus handles agent status change from runner
func (pc *PodCoordinator) handleAgentStatus(runnerID int64, data *AgentStatusData) {
	ctx := context.Background()

	updates := map[string]interface{}{
		"agent_status": data.Status,
	}
	if data.Pid > 0 {
		updates["pty_pid"] = data.Pid
	}

	if err := pc.db.WithContext(ctx).
		Model(&agentpod.Pod{}).
		Where("pod_key = ?", data.PodKey).
		Updates(updates).Error; err != nil {
		pc.logger.Error("failed to update agent status",
			"pod_key", data.PodKey,
			"error", err)
		return
	}

	pc.logger.Debug("agent status changed",
		"pod_key", data.PodKey,
		"status", data.Status)

	// Notify status change
	if pc.onStatusChange != nil {
		pc.onStatusChange(data.PodKey, "", data.Status)
	}
}

// handleRunnerDisconnect handles runner disconnection
func (pc *PodCoordinator) handleRunnerDisconnect(runnerID int64) {
	ctx := context.Background()

	// Mark runner as offline, but don't immediately orphan pods
	// Pods will be orphaned by reconcilePods if runner doesn't reconnect
	// and report them in heartbeat
	if err := pc.db.WithContext(ctx).
		Table("runners").
		Where("id = ?", runnerID).
		Update("status", "offline").Error; err != nil {
		pc.logger.Error("failed to mark runner as offline",
			"runner_id", runnerID,
			"error", err)
	}

	pc.logger.Info("runner disconnected, pods will be reconciled on reconnect",
		"runner_id", runnerID)

	// Note: We intentionally don't mark pods as orphaned here
	// The runner might reconnect quickly (network glitch) and pods are still running
	// Pods will be properly reconciled when:
	// 1. Runner reconnects and sends heartbeat - reconcilePods will handle it
	// 2. Pod cleanup task runs and finds stale pods
}

// IncrementPods increments pod count for a runner
func (pc *PodCoordinator) IncrementPods(ctx context.Context, runnerID int64) error {
	return pc.db.WithContext(ctx).Exec(
		"UPDATE runners SET current_pods = current_pods + 1 WHERE id = ?",
		runnerID,
	).Error
}

// DecrementPods decrements pod count for a runner
func (pc *PodCoordinator) DecrementPods(ctx context.Context, runnerID int64) error {
	return pc.db.WithContext(ctx).Exec(
		"UPDATE runners SET current_pods = GREATEST(current_pods - 1, 0) WHERE id = ?",
		runnerID,
	).Error
}

// CreatePod creates a new pod on a runner
func (pc *PodCoordinator) CreatePod(ctx context.Context, runnerID int64, req *CreatePodRequest) error {
	// Increment pod count first
	if err := pc.IncrementPods(ctx, runnerID); err != nil {
		return err
	}

	// Register with terminal router
	pc.terminalRouter.RegisterPod(req.PodKey, runnerID)

	// Send create pod request to runner
	return pc.connectionManager.SendCreatePod(ctx, runnerID, req)
}

// TerminatePod terminates a pod on a runner
func (pc *PodCoordinator) TerminatePod(ctx context.Context, podKey string) error {
	// Get pod to find runner
	var pod agentpod.Pod
	if err := pc.db.WithContext(ctx).
		Where("pod_key = ?", podKey).
		First(&pod).Error; err != nil {
		return err
	}

	// Send terminate request to runner
	if err := pc.connectionManager.SendTerminatePod(ctx, pod.RunnerID, podKey); err != nil {
		pc.logger.Warn("failed to send terminate to runner, marking as completed",
			"pod_key", podKey,
			"error", err)
	}

	// Update pod status
	now := time.Now()
	if err := pc.db.WithContext(ctx).
		Model(&pod).
		Updates(map[string]interface{}{
			"status":      agentpod.StatusCompleted,
			"finished_at": now,
		}).Error; err != nil {
		return err
	}

	// Unregister from terminal router
	pc.terminalRouter.UnregisterPod(podKey)

	// Decrement pod count
	return pc.DecrementPods(ctx, pod.RunnerID)
}

// UpdateActivity updates last activity timestamp for a pod
func (pc *PodCoordinator) UpdateActivity(ctx context.Context, podKey string) error {
	return pc.db.WithContext(ctx).
		Model(&agentpod.Pod{}).
		Where("pod_key = ?", podKey).
		Update("last_activity", time.Now()).Error
}

// MarkDisconnected marks a pod as disconnected (user closed browser)
func (pc *PodCoordinator) MarkDisconnected(ctx context.Context, podKey string) error {
	return pc.db.WithContext(ctx).
		Model(&agentpod.Pod{}).
		Where("pod_key = ? AND status = ?", podKey, agentpod.StatusRunning).
		Update("status", agentpod.StatusDisconnected).Error
}

// MarkReconnected marks a pod as running again (user reconnected)
func (pc *PodCoordinator) MarkReconnected(ctx context.Context, podKey string) error {
	return pc.db.WithContext(ctx).
		Model(&agentpod.Pod{}).
		Where("pod_key = ? AND status = ?", podKey, agentpod.StatusDisconnected).
		Updates(map[string]interface{}{
			"status":        agentpod.StatusRunning,
			"last_activity": time.Now(),
		}).Error
}

// hashCapabilities computes a SHA256 hash of capabilities for change detection.
// This avoids unnecessary database updates when capabilities haven't changed.
func (pc *PodCoordinator) hashCapabilities(caps []runnerDomain.PluginCapability) string {
	data, err := json.Marshal(caps)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}
