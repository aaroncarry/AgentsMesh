package v1

import (
	"github.com/anthropics/agentsmesh/backend/internal/service/billing"
	grantservice "github.com/anthropics/agentsmesh/backend/internal/service/grant"
	"github.com/anthropics/agentsmesh/backend/internal/service/repository"
)

// RepositoryHandler handles repository-related requests
type RepositoryHandler struct {
	repositoryService repository.RepositoryServiceInterface
	billingService    *billing.Service
	grantService      *grantservice.Service
}

// NewRepositoryHandler creates a new repository handler
func NewRepositoryHandler(repositoryService repository.RepositoryServiceInterface, opts ...RepositoryHandlerOption) *RepositoryHandler {
	h := &RepositoryHandler{
		repositoryService: repositoryService,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// RepositoryHandlerOption is a functional option for configuring RepositoryHandler
type RepositoryHandlerOption func(*RepositoryHandler)

// WithBillingService sets the billing service for quota checks
func WithBillingService(bs *billing.Service) RepositoryHandlerOption {
	return func(h *RepositoryHandler) {
		h.billingService = bs
	}
}

// WithGrantServiceForRepo sets the grant service for resource sharing
func WithGrantServiceForRepo(gs *grantservice.Service) RepositoryHandlerOption {
	return func(h *RepositoryHandler) {
		h.grantService = gs
	}
}

// CreateRepositoryRequest represents repository creation request
// Self-contained: includes all provider info, no git_provider_id
type CreateRepositoryRequest struct {
	ProviderType    string `json:"provider_type" binding:"required"`     // github, gitlab, gitee, generic
	ProviderBaseURL string `json:"provider_base_url" binding:"required"` // https://github.com, https://gitlab.company.com
	HttpCloneURL    string `json:"http_clone_url"`                       // HTTPS clone URL (optional, will be generated)
	SshCloneURL     string `json:"ssh_clone_url"`                        // SSH clone URL (optional, will be generated)
	ExternalID      string `json:"external_id" binding:"required"`
	Name            string `json:"name" binding:"required"`
	Slug            string `json:"slug" binding:"required"`
	DefaultBranch   string `json:"default_branch"`
	TicketPrefix    string `json:"ticket_prefix"`
	Visibility      string `json:"visibility"` // "organization" or "private", defaults to "organization"
}

// UpdateRepositoryRequest represents repository update request
type UpdateRepositoryRequest struct {
	Name          string  `json:"name"`
	DefaultBranch string  `json:"default_branch"`
	TicketPrefix  string  `json:"ticket_prefix"`
	IsActive      *bool   `json:"is_active"`
	HttpCloneURL  *string `json:"http_clone_url"`
	SshCloneURL   *string `json:"ssh_clone_url"`
}

// SyncBranchesRequest represents sync branches request
type SyncBranchesRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
}
