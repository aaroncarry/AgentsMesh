package v1

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/anthropics/agentsmesh/backend/pkg/policy"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpgradeRunnerRequest represents the request body for runner upgrade.
type UpgradeRunnerRequest struct {
	TargetVersion string `json:"target_version"`
	// Deprecated: Force is accepted for backward compatibility but ignored.
	// The upgrade path is always forced now that Poddaemon keeps pods alive
	// across Runner restarts.
	Force bool `json:"force,omitempty"`
}

// UpgradeRunner triggers a remote upgrade on a runner.
// POST /api/v1/organizations/:slug/runners/:id/upgrade
func (h *RunnerHandler) UpgradeRunner(c *gin.Context) {
	if h.upgradeCommandSender == nil {
		apierr.ServiceUnavailable(c, apierr.SERVICE_UNAVAILABLE, "Upgrade service not configured")
		return
	}

	runnerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apierr.InvalidInput(c, "Invalid runner ID")
		return
	}

	var req UpgradeRunnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body (upgrade to latest)
		req = UpgradeRunnerRequest{}
	}

	tenant := middleware.GetTenant(c)
	sub := policy.NewSubject(tenant.OrganizationID, tenant.UserID, tenant.UserRole)

	r, err := h.runnerService.GetRunner(c.Request.Context(), runnerID)
	if err != nil {
		apierr.ResourceNotFound(c, "Runner not found")
		return
	}

	// Upgrade triggers a binary download + process restart — align with
	// Update/Delete (AllowWrite) rather than the read-only permission previously
	// used here. Read-only grants must not be able to force a runner upgrade.
	if !policy.RunnerPolicy.AllowWrite(sub, policy.VisibleResource(
		r.OrganizationID, r.RegisteredByUserID, r.Visibility,
	)) {
		apierr.ForbiddenAccess(c)
		return
	}

	if req.Force {
		slog.WarnContext(c.Request.Context(), "Deprecated 'force' field received — ignored since Poddaemon upgrade path",
			"runner_id", runnerID, "user_id", tenant.UserID)
	}

	// Check if runner is online
	if !h.upgradeCommandSender.IsConnected(runnerID) {
		apierr.ServiceUnavailable(c, apierr.SERVICE_UNAVAILABLE, "Runner is not connected")
		return
	}

	// Generate request ID and send upgrade command
	// NOTE: force=true always — Poddaemon ensures pods survive runner restarts,
	// so the old pod-count guard is no longer needed.
	requestID := uuid.New().String()
	if err := h.upgradeCommandSender.SendUpgradeRunner(runnerID, requestID, req.TargetVersion, true); err != nil {
		// Differentiate error types for better client diagnostics
		if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
			apierr.ServiceUnavailable(c, apierr.SERVICE_UNAVAILABLE, "Runner disconnected before command could be sent")
		} else {
			apierr.InternalError(c, "Failed to send upgrade command")
		}
		return
	}

	// Audit log — active_pod_count is recorded so post-incident analysis can
	// correlate an upgrade with any user sessions that may have been affected.
	slog.InfoContext(c.Request.Context(), "Runner upgrade initiated",
		"runner_id", runnerID,
		"request_id", requestID,
		"target_version", req.TargetVersion,
		"active_pod_count", r.CurrentPods,
		"user_id", tenant.UserID,
		"org_id", tenant.OrganizationID,
	)

	c.JSON(http.StatusAccepted, gin.H{
		"request_id": requestID,
		"message":    "Upgrade command sent to runner",
	})
}
