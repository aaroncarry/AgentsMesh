package v1

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/grant"
	"github.com/anthropics/agentsmesh/backend/pkg/policy"
)

// podResourceWithGrants builds a ResourceContext for a pod, including grants if available.
func (h *PodHandler) podResourceWithGrants(ctx context.Context, podKey string, orgID, createdByID int64) policy.ResourceContext {
	rc := policy.PodResource(orgID, createdByID)
	if h.grantService == nil {
		return rc
	}
	if ids, err := h.grantService.GetGrantedUserIDs(ctx, grant.TypePod, podKey); err == nil && len(ids) > 0 {
		return rc.WithGrants(ids)
	}
	return rc
}

// runnerResourceWithGrants builds a ResourceContext for a runner, including grants if available.
func (h *RunnerHandler) runnerResourceWithGrants(ctx context.Context, runnerID int64, orgID int64, registeredByUserID *int64, visibility string) policy.ResourceContext {
	rc := policy.VisibleResource(orgID, registeredByUserID, visibility)
	if h.grantService == nil {
		return rc
	}
	if ids, err := h.grantService.GetGrantedUserIDs(ctx, grant.TypeRunner, grant.IntResourceID(runnerID)); err == nil && len(ids) > 0 {
		return rc.WithGrants(ids)
	}
	return rc
}

// repoResourceWithGrants builds a ResourceContext for a repository, including grants if available.
func (h *RepositoryHandler) repoResourceWithGrants(ctx context.Context, repoID int64, orgID int64, importedByUserID *int64, visibility string) policy.ResourceContext {
	rc := policy.VisibleResource(orgID, importedByUserID, visibility)
	if h.grantService == nil {
		return rc
	}
	if ids, err := h.grantService.GetGrantedUserIDs(ctx, grant.TypeRepository, grant.IntResourceID(repoID)); err == nil && len(ids) > 0 {
		return rc.WithGrants(ids)
	}
	return rc
}
