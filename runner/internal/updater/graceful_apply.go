package updater

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/process"
)

// executeUpdate creates a backup, downloads and replaces the binary, and restarts.
func (g *GracefulUpdater) executeUpdate(ctx context.Context) error {
	g.mu.RLock()
	info := g.pendingInfo
	g.mu.RUnlock()

	if info == nil {
		g.setState(StateIdle)
		return fmt.Errorf("no pending update to apply")
	}

	g.setState(StateApplying)

	// Create backup for potential rollback
	backupPath, err := g.updater.CreateBackup()
	if err != nil {
		log.Printf("[updater] Warning: failed to create backup: %v", err)
		// Continue without backup - rollback won't be possible
	}

	// Update binary in-place via detector.UpdateBinary
	if err := g.updater.updateBinary(ctx, info.LatestVersion); err != nil {
		g.mu.Lock()
		g.pendingInfo = nil
		g.mu.Unlock()
		g.setState(StateIdle)
		return fmt.Errorf("failed to apply update: %w", err)
	}

	g.mu.Lock()
	g.pendingInfo = nil
	g.mu.Unlock()

	log.Printf("[updater] Update applied successfully: %s -> %s", info.CurrentVersion, info.LatestVersion)

	// Restart
	g.setState(StateRestarting)
	if g.restartFunc != nil {
		pid, err := g.restartFunc()
		if err != nil {
			// The binary on disk is already updated. If we can't restart
			// in-process (e.g., /proc/self/exe points to a deleted .old file),
			// exit so the service manager (systemd) restarts us with the new
			// binary. Rolling back here would leave the old version running
			// permanently since subsequent upgrade attempts also fail once
			// /proc/self/exe is stale.
			log.Printf("[updater] Restart failed after successful update: %v", err)
			log.Printf("[updater] Exiting so the service manager restarts with the new binary")
			g.exitFunc(0)
		}

		// Health check if configured
		if g.healthChecker != nil && pid > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), g.healthTimeout)
			defer cancel()

			if err := g.healthChecker(ctx, pid); err != nil {
				log.Printf("[updater] Health check failed, attempting rollback: %v", err)
				// Terminate the unhealthy new process
				if proc, findErr := os.FindProcess(pid); findErr == nil && proc != nil {
					_ = proc.Kill()
				}
				if rbErr := g.rollbackUpdate(backupPath); rbErr != nil {
					log.Printf("[updater] Rollback also failed: %v", rbErr)
				}
				g.setState(StateIdle)
				return fmt.Errorf("health check failed: %w", err)
			}
			log.Printf("[updater] Health check passed for new process (PID: %d)", pid)
		}
	}

	return nil
}

// rollbackUpdate attempts to restore the previous version from backup.
func (g *GracefulUpdater) rollbackUpdate(backupPath string) error {
	if backupPath == "" {
		return fmt.Errorf("no backup available for rollback")
	}
	if err := g.updater.Rollback(); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}
	log.Printf("[updater] Successfully rolled back to previous version")
	return nil
}

// DefaultRestartFunc returns a restart function that re-executes the current binary.
// Note: This function starts a new process and signals the caller to exit gracefully.
// The caller should handle process termination appropriately.
// Returns the PID of the new process for health checking.
func DefaultRestartFunc() RestartFunc {
	return func() (int, error) {
		execPath, err := os.Executable()
		if err != nil {
			return 0, fmt.Errorf("failed to get executable path: %w", err)
		}

		// Start new process
		cmd := exec.Command(execPath, os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = os.Environ()

		if err := cmd.Start(); err != nil {
			return 0, fmt.Errorf("failed to start new process: %w", err)
		}

		log.Printf("[updater] New process started (PID: %d), current process should exit", cmd.Process.Pid)
		// Note: Caller is responsible for graceful shutdown after this returns
		// Do NOT call os.Exit() here as it prevents proper cleanup
		return cmd.Process.Pid, nil
	}
}

// DefaultHealthChecker returns a health checker that validates the new process is running.
// minRunTime: the minimum time the new process should run before being considered healthy.
func DefaultHealthChecker(minRunTime time.Duration) HealthChecker {
	return func(ctx context.Context, pid int) error {
		// Wait for the specified minimum run time
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(minRunTime):
		}

		// Check if the process is still running (cross-platform)
		return process.IsAlive(pid)
	}
}
