package middleware

import (
	"context"
	"net/http"

	"github.com/anthropics/agentmesh/backend/internal/domain/agentpod"
	"github.com/gin-gonic/gin"
)

// PodService interface for pod lookup
type PodService interface {
	GetPodByKey(ctx context.Context, podKey string) (*agentpod.Pod, error)
}

// PodAuthMiddleware extracts pod key from X-Pod-Key header
// and sets up the tenant context based on the pod's organization.
// This allows MCP tools to access organization-scoped APIs without
// requiring the organization slug in the URL.
func PodAuthMiddleware(podService PodService, orgService OrganizationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		podKey := c.GetHeader("X-Pod-Key")
		if podKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "X-Pod-Key header required",
			})
			c.Abort()
			return
		}

		// Get pod by key
		pod, err := podService.GetPodByKey(c.Request.Context(), podKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid pod key",
			})
			c.Abort()
			return
		}

		orgID := pod.OrganizationID

		// Create tenant context with pod info
		// Use pod's CreatedByID as the user ID for permission checks
		// This ensures MCP tools operate with the pod creator's permissions
		tc := &TenantContext{
			OrganizationID:   orgID,
			OrganizationSlug: "", // Will be filled if needed
			UserID:           pod.CreatedByID, // Use pod creator's ID
			UserRole:         "pod", // Special role for pod-based access
		}

		// Store pod key in context for later use
		c.Set("pod_key", podKey)
		c.Set("tenant", tc)
		ctx := SetTenant(c.Request.Context(), tc)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
