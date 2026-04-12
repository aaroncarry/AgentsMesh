package v1

import (
	grantservice "github.com/anthropics/agentsmesh/backend/internal/service/grant"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

type grantAccessRequest struct {
	UserID int64 `json:"user_id" binding:"required"`
}

func handleGrantError(c *gin.Context, err error) {
	switch err {
	case grantservice.ErrSelfGrant:
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, "Cannot grant access to yourself")
	case grantservice.ErrInvalidType:
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, "Invalid resource type")
	default:
		apierr.InternalError(c, "Failed to grant access")
	}
}
