package v1

import (
	"net/http"

	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

// ArchiveChannel archives a channel
// POST /api/v1/organizations/:slug/channels/:id/archive
func (h *ChannelHandler) ArchiveChannel(c *gin.Context) {
	ch, ok := h.requireChannelAccess(c)
	if !ok {
		return
	}

	if err := h.channelService.ArchiveChannel(c.Request.Context(), ch.ID); err != nil {
		apierr.InternalError(c, "Failed to archive channel")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Channel archived"})
}

// UnarchiveChannel unarchives a channel
// POST /api/v1/organizations/:slug/channels/:id/unarchive
func (h *ChannelHandler) UnarchiveChannel(c *gin.Context) {
	ch, ok := h.requireChannelAccess(c)
	if !ok {
		return
	}

	if err := h.channelService.UnarchiveChannel(c.Request.Context(), ch.ID); err != nil {
		apierr.InternalError(c, "Failed to unarchive channel")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Channel unarchived"})
}
