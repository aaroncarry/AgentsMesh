package middleware

import (
	"context"
	"net/http"

	"github.com/anthropics/agentmesh/backend/internal/domain/session"
	"github.com/gin-gonic/gin"
)

// SessionService interface for session lookup
type SessionService interface {
	GetSessionByKey(ctx context.Context, sessionKey string) (*session.Session, error)
}

// SessionAuthMiddleware extracts session key from X-Session-Key header
// and sets up the tenant context based on the session's organization.
// This allows MCP tools to access organization-scoped APIs without
// requiring the organization slug in the URL.
func SessionAuthMiddleware(sessionService SessionService, orgService OrganizationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionKey := c.GetHeader("X-Session-Key")
		if sessionKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "X-Session-Key header required",
			})
			c.Abort()
			return
		}

		// Get session by key
		sess, err := sessionService.GetSessionByKey(c.Request.Context(), sessionKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid session key",
			})
			c.Abort()
			return
		}

		orgID := sess.OrganizationID

		// Create tenant context with session info
		// Use session's CreatedByID as the user ID for permission checks
		// This ensures MCP tools operate with the session creator's permissions
		tc := &TenantContext{
			OrganizationID:   orgID,
			OrganizationSlug: "", // Will be filled if needed
			UserID:           sess.CreatedByID, // Use session creator's ID
			UserRole:         "session", // Special role for session-based access
		}

		// Store session key in context for later use
		c.Set("session_key", sessionKey)
		c.Set("tenant", tc)
		ctx := SetTenant(c.Request.Context(), tc)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
