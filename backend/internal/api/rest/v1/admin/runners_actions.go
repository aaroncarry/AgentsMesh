package admin

import (
	"net/http"
	"strconv"

	"github.com/anthropics/agentsmesh/backend/internal/domain/admin"
	adminservice "github.com/anthropics/agentsmesh/backend/internal/service/admin"

	"github.com/gin-gonic/gin"
)

// DisableRunner disables a runner
func (h *RunnerHandler) DisableRunner(c *gin.Context) {
	runnerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner ID"})
		return
	}

	// Get old data for audit log
	oldRunner, _ := h.adminService.GetRunner(c.Request.Context(), runnerID)

	r, err := h.adminService.DisableRunner(c.Request.Context(), runnerID)
	if err != nil {
		if err == adminservice.ErrRunnerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable runner"})
		return
	}

	// Log disable action
	h.logAction(c, admin.AuditActionRunnerDisable, admin.TargetTypeRunner, runnerID, oldRunner, r)

	c.JSON(http.StatusOK, runnerResponse(r))
}

// EnableRunner enables a runner
func (h *RunnerHandler) EnableRunner(c *gin.Context) {
	runnerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner ID"})
		return
	}

	// Get old data for audit log
	oldRunner, _ := h.adminService.GetRunner(c.Request.Context(), runnerID)

	r, err := h.adminService.EnableRunner(c.Request.Context(), runnerID)
	if err != nil {
		if err == adminservice.ErrRunnerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable runner"})
		return
	}

	// Log enable action
	h.logAction(c, admin.AuditActionRunnerEnable, admin.TargetTypeRunner, runnerID, oldRunner, r)

	c.JSON(http.StatusOK, runnerResponse(r))
}

// DeleteRunner deletes a runner
func (h *RunnerHandler) DeleteRunner(c *gin.Context) {
	runnerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner ID"})
		return
	}

	deletedRunner, err := h.adminService.DeleteRunner(c.Request.Context(), runnerID)
	if err != nil {
		if err == adminservice.ErrRunnerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
			return
		}
		if err == adminservice.ErrRunnerHasActivePods {
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete runner with active pods"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete runner"})
		return
	}

	// Log delete action
	h.logAction(c, admin.AuditActionRunnerDelete, admin.TargetTypeRunner, runnerID, deletedRunner, nil)

	c.JSON(http.StatusOK, gin.H{"message": "Runner deleted successfully"})
}
