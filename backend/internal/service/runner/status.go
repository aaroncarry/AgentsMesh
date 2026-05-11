package runner

import (
	"context"
	"log/slog"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/runner"
)

// UpdateRunnerStatus updates runner status
func (s *Service) UpdateRunnerStatus(ctx context.Context, runnerID int64, status string) error {
	now := time.Now()
	return s.repo.UpdateFields(ctx, runnerID, map[string]interface{}{
		"status":         status,
		"last_heartbeat": now,
	})
}

// SetRunnerStatus sets the runner status (alias for UpdateRunnerStatus)
func (s *Service) SetRunnerStatus(ctx context.Context, runnerID int64, status string) error {
	return s.UpdateRunnerStatus(ctx, runnerID, status)
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
		slog.ErrorContext(ctx, "failed to get runner for connect", "runner_id", runnerID, "error", err)
		return err
	}

	// Update status in the Runner object before caching
	r.Status = runner.RunnerStatusOnline

	now := time.Now()
	s.activeRunners.Store(runnerID, &ActiveRunner{
		Runner:   r,
		LastPing: now,
		PodCount: r.CurrentPods,
	})

	slog.InfoContext(ctx, "runner connected", "runner_id", runnerID)
	return s.UpdateRunnerStatus(ctx, runnerID, runner.RunnerStatusOnline)
}

// MarkDisconnected marks a runner as disconnected
func (s *Service) MarkDisconnected(ctx context.Context, runnerID int64) error {
	s.activeRunners.Delete(runnerID)
	slog.InfoContext(ctx, "runner disconnected", "runner_id", runnerID)
	return s.UpdateRunnerStatus(ctx, runnerID, runner.RunnerStatusOffline)
}

// UpdateHostInfo updates runner host information
func (s *Service) UpdateHostInfo(ctx context.Context, runnerID int64, hostInfo map[string]interface{}) error {
	return s.repo.UpdateFields(ctx, runnerID, map[string]interface{}{
		"host_info": hostInfo,
	})
}

// UpdateRunnerVersionAndHostInfo updates runner version and host information atomically.
// Called during gRPC initialization handshake to persist RunnerInfo from the connect request.
func (s *Service) UpdateRunnerVersionAndHostInfo(ctx context.Context, runnerID int64, version string, hostInfo map[string]interface{}) error {
	return s.repo.UpdateFields(ctx, runnerID, map[string]interface{}{
		"runner_version": version,
		"host_info":      hostInfo,
	})
}

// UpdateAvailableAgents updates the list of available agents for a runner
// Called when runner completes initialization handshake
func (s *Service) UpdateAvailableAgents(ctx context.Context, runnerID int64, agents []string) error {
	agents = normalizeAgentSlugs(agents)
	slog.InfoContext(ctx, "runner available agents updated", "runner_id", runnerID, "agents", agents)
	if err := s.repo.UpdateFields(ctx, runnerID, map[string]interface{}{
		"available_agents": runner.StringSlice(agents),
	}); err != nil {
		return err
	}

	if active, ok := s.activeRunners.Load(runnerID); ok {
		if ar, ok := active.(*ActiveRunner); ok && ar.Runner != nil {
			updated := *ar.Runner
			updated.AvailableAgents = runner.StringSlice(agents)
			s.activeRunners.Store(runnerID, &ActiveRunner{
				Runner:   &updated,
				LastPing: ar.LastPing,
				PodCount: ar.PodCount,
			})
		}
	}
	return nil
}

// UpdateAgentVersions updates the detected agent version info for a runner.
// Called when runner completes initialization handshake (Runner >= 0.4.7).
// Also refreshes the activeRunners cache to keep GetRunner consistent.
func (s *Service) UpdateAgentVersions(ctx context.Context, runnerID int64, versions []runner.AgentVersion) error {
	versions = normalizeAgentVersions(versions)
	if err := s.repo.UpdateFields(ctx, runnerID, map[string]interface{}{
		"agent_versions": runner.AgentVersionSlice(versions),
	}); err != nil {
		return err
	}

	// Sync in-memory cache so GetRunner returns fresh data immediately.
	if active, ok := s.activeRunners.Load(runnerID); ok {
		if ar, ok := active.(*ActiveRunner); ok && ar.Runner != nil {
			updated := *ar.Runner
			updated.AgentVersions = runner.AgentVersionSlice(versions)
			s.activeRunners.Store(runnerID, &ActiveRunner{
				Runner:   &updated,
				LastPing: ar.LastPing,
				PodCount: ar.PodCount,
			})
		}
	}
	return nil
}

// MergeAgentVersions merges delta agent version updates into existing versions.
// Entries where both Version and Path are empty are treated as removals (agent no longer available).
// Called when runner reports version changes via heartbeat.
//
// NOTE: This method uses read-modify-write and is NOT safe for concurrent calls on the same runner.
// Correctness relies on gRPC recvLoop serializing all heartbeat messages per runner.
func (s *Service) MergeAgentVersions(ctx context.Context, runnerID int64, changes map[string]runner.AgentVersion) error {
	if len(changes) == 0 {
		return nil
	}

	r, err := s.GetRunner(ctx, runnerID)
	if err != nil {
		return err
	}

	// Build merged version map from existing data
	merged := make(map[string]runner.AgentVersion)
	for _, v := range r.AgentVersions {
		v.Slug = canonicalAgentSlug(v.Slug)
		merged[v.Slug] = v
	}

	// Apply changes
	for slug, change := range changes {
		slug = canonicalAgentSlug(slug)
		change.Slug = canonicalAgentSlug(change.Slug)
		if change.Version == "" && change.Path == "" {
			delete(merged, slug)
		} else {
			merged[slug] = change
		}
	}

	// Convert back to slice
	result := make([]runner.AgentVersion, 0, len(merged))
	for _, v := range merged {
		result = append(result, v)
	}

	return s.UpdateAgentVersions(ctx, runnerID, result)
}

func normalizeAgentSlugs(agents []string) []string {
	if len(agents) == 0 {
		return agents
	}

	normalized := make([]string, 0, len(agents))
	seen := make(map[string]struct{}, len(agents))
	for _, slug := range agents {
		slug = canonicalAgentSlug(slug)
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}
		normalized = append(normalized, slug)
	}
	return normalized
}

func normalizeAgentVersions(versions []runner.AgentVersion) []runner.AgentVersion {
	if len(versions) == 0 {
		return versions
	}

	normalized := make([]runner.AgentVersion, 0, len(versions))
	indexBySlug := make(map[string]int, len(versions))
	for _, version := range versions {
		version.Slug = canonicalAgentSlug(version.Slug)
		if idx, ok := indexBySlug[version.Slug]; ok {
			normalized[idx] = version
			continue
		}
		indexBySlug[version.Slug] = len(normalized)
		normalized = append(normalized, version)
	}
	return normalized
}

func canonicalAgentSlug(slug string) string {
	switch slug {
	case "factory-droid":
		return "factory-cli"
	default:
		return slug
	}
}

// IncrementPods increments the pod count for a runner
func (s *Service) IncrementPods(ctx context.Context, runnerID int64) error {
	return s.repo.IncrementPods(ctx, runnerID)
}

// DecrementPods decrements the pod count for a runner
func (s *Service) DecrementPods(ctx context.Context, runnerID int64) error {
	return s.repo.DecrementPods(ctx, runnerID)
}

// RunnerUpdateFunc is a callback for runner status updates
type RunnerUpdateFunc func(*runner.Runner)

// SubscribeStatusChanges subscribes to runner status changes and returns an unsubscribe function
func (s *Service) SubscribeStatusChanges(ctx context.Context, callback RunnerUpdateFunc) (func(), error) {
	// In a real implementation, this would use Redis pub/sub or similar
	// For now, return a simple unsubscribe function
	return func() {}, nil
}
