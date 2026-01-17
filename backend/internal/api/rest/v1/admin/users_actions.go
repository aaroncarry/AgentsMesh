package admin

import (
	"net/http"
	"strconv"

	"github.com/anthropics/agentsmesh/backend/internal/domain/admin"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	adminservice "github.com/anthropics/agentsmesh/backend/internal/service/admin"

	"github.com/gin-gonic/gin"
)

// DisableUser disables a user account
func (h *UserHandler) DisableUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Prevent disabling self
	adminUserID := middleware.GetAdminUserID(c)
	if userID == adminUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot disable your own account"})
		return
	}

	// Get old data for audit log
	oldUser, _ := h.adminService.GetUser(c.Request.Context(), userID)

	user, err := h.adminService.DisableUser(c.Request.Context(), userID)
	if err != nil {
		if err == adminservice.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable user"})
		return
	}

	// Log disable action
	h.logAction(c, admin.AuditActionUserDisable, admin.TargetTypeUser, userID, oldUser, user)

	c.JSON(http.StatusOK, adminUserResponse(user))
}

// EnableUser enables a user account
func (h *UserHandler) EnableUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get old data for audit log
	oldUser, _ := h.adminService.GetUser(c.Request.Context(), userID)

	user, err := h.adminService.EnableUser(c.Request.Context(), userID)
	if err != nil {
		if err == adminservice.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable user"})
		return
	}

	// Log enable action
	h.logAction(c, admin.AuditActionUserEnable, admin.TargetTypeUser, userID, oldUser, user)

	c.JSON(http.StatusOK, adminUserResponse(user))
}

// GrantAdmin grants system admin privileges to a user
func (h *UserHandler) GrantAdmin(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get old data for audit log
	oldUser, _ := h.adminService.GetUser(c.Request.Context(), userID)

	user, err := h.adminService.GrantAdmin(c.Request.Context(), userID)
	if err != nil {
		if err == adminservice.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to grant admin privileges"})
		return
	}

	// Log grant admin action
	h.logAction(c, admin.AuditActionUserGrantAdmin, admin.TargetTypeUser, userID, oldUser, user)

	c.JSON(http.StatusOK, adminUserResponse(user))
}

// RevokeAdmin revokes system admin privileges from a user
func (h *UserHandler) RevokeAdmin(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	adminUserID := middleware.GetAdminUserID(c)

	// Get old data for audit log
	oldUser, _ := h.adminService.GetUser(c.Request.Context(), userID)

	user, err := h.adminService.RevokeAdmin(c.Request.Context(), userID, adminUserID)
	if err != nil {
		if err == adminservice.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if err == adminservice.ErrCannotRevokeOwnAdmin {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot revoke your own admin privileges"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke admin privileges"})
		return
	}

	// Log revoke admin action
	h.logAction(c, admin.AuditActionUserRevokeAdmin, admin.TargetTypeUser, userID, oldUser, user)

	c.JSON(http.StatusOK, adminUserResponse(user))
}
