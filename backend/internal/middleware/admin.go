package middleware

import (
	"net/http"

	"github.com/anthropics/agentsmesh/backend/internal/domain/user"
	"github.com/anthropics/agentsmesh/backend/internal/infra/database"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware validates that the authenticated user is a system admin.
// This middleware must be used after AuthMiddleware.
func AdminMiddleware(db database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by AuthMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Fetch user from database to verify is_system_admin flag
		var u user.User
		if err := db.First(&u, userID); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found",
			})
			c.Abort()
			return
		}

		// Verify user is a system admin
		if !u.IsSystemAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied: system administrator privileges required",
			})
			c.Abort()
			return
		}

		// Verify user is active
		if !u.IsActive {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied: user account is disabled",
			})
			c.Abort()
			return
		}

		// Set admin user in context for audit logging
		c.Set("admin_user", &u)
		c.Set("admin_user_id", u.ID)

		c.Next()
	}
}

// GetAdminUser retrieves the admin user from the context
func GetAdminUser(c *gin.Context) *user.User {
	if u, exists := c.Get("admin_user"); exists {
		if adminUser, ok := u.(*user.User); ok {
			return adminUser
		}
	}
	return nil
}

// GetAdminUserID retrieves the admin user ID from the context
func GetAdminUserID(c *gin.Context) int64 {
	if id, exists := c.Get("admin_user_id"); exists {
		if adminID, ok := id.(int64); ok {
			return adminID
		}
	}
	return 0
}
