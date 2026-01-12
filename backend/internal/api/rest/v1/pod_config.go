package v1

import (
	"log"

	"github.com/anthropics/agentmesh/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// buildPluginConfig builds the PluginConfig map for Runner's Sandbox plugins
// It resolves repository_id -> repository_url, ticket_id -> ticket_identifier
//
// Configuration Priority (later overrides earlier):
// 1. System defaults (hardcoded)
// 2. Organization default config (from organization_agent_configs)
// 3. User request plugin_config
//
// Credential Strategy (权限跟人走):
// - Repository is self-contained with provider_type and provider_base_url
// - Credentials are obtained from the current user's default Git Credential
// - If no user credentials available or using runner_local, Runner uses its local Git configuration
func (h *PodHandler) buildPluginConfig(c *gin.Context, req *CreatePodRequest) map[string]interface{} {
	config := make(map[string]interface{})
	ctx := c.Request.Context()
	tenant := middleware.GetTenant(c)
	userID := middleware.GetUserID(c)

	// 1. Get organization default config for the agent type
	if req.AgentTypeID != nil && h.agentService != nil {
		orgConfig := h.agentService.GetEffectiveConfig(ctx, tenant.OrganizationID, *req.AgentTypeID, nil)
		for k, v := range orgConfig {
			config[k] = v
		}
	}

	// 2. Resolve Repository URL
	h.resolveRepositoryConfig(c, req, config)

	// 3. Get Git credentials from current user (权限跟人走)
	h.resolveGitCredentials(c, userID, config)

	// 4. Resolve Branch Name
	h.resolveBranchConfig(c, req, config)

	// 5. Resolve Ticket Identifier
	h.resolveTicketConfig(c, req, config)

	// 6. Merge user-provided PluginConfig (can override above values)
	if req.PluginConfig != nil {
		for k, v := range req.PluginConfig {
			config[k] = v
		}
	}

	// 7. Inject agent credentials as environment variables (权限跟人走)
	h.resolveAgentCredentials(c, req, userID, config)

	return config
}

// resolveRepositoryConfig resolves repository URL from request
// Priority: repository_url > repository_id
func (h *PodHandler) resolveRepositoryConfig(c *gin.Context, req *CreatePodRequest, config map[string]interface{}) {
	ctx := c.Request.Context()

	if req.RepositoryURL != nil && *req.RepositoryURL != "" {
		config["repository_url"] = *req.RepositoryURL
	} else if req.RepositoryID != nil && h.repositoryService != nil {
		repo, err := h.repositoryService.GetByID(ctx, *req.RepositoryID)
		if err == nil && repo != nil {
			// Use repository's self-contained clone URL
			config["repository_url"] = repo.CloneURL

			// Store provider info for potential use by Runner
			config["provider_type"] = repo.ProviderType
			config["provider_base_url"] = repo.ProviderBaseURL
		}
	}
}

// resolveGitCredentials resolves Git credentials from current user (权限跟人走)
func (h *PodHandler) resolveGitCredentials(c *gin.Context, userID int64, config map[string]interface{}) {
	if h.userService == nil {
		return
	}

	gitCred := h.getUserGitCredential(c, userID)
	if gitCred == nil {
		return
	}

	switch gitCred.Type {
	case "oauth", "pat":
		if gitCred.Token != "" {
			config["git_token"] = gitCred.Token
		}
	case "ssh_key":
		if gitCred.SSHPrivateKey != "" {
			config["ssh_private_key"] = gitCred.SSHPrivateKey
		}
	// case "runner_local": no credentials needed, Runner uses local config
	}
}

// resolveBranchConfig resolves branch name from request or repository default
func (h *PodHandler) resolveBranchConfig(c *gin.Context, req *CreatePodRequest, config map[string]interface{}) {
	ctx := c.Request.Context()

	if req.BranchName != nil && *req.BranchName != "" {
		config["branch"] = *req.BranchName
	} else if req.RepositoryID != nil && h.repositoryService != nil {
		// Use repository's default branch if not specified
		repo, err := h.repositoryService.GetByID(ctx, *req.RepositoryID)
		if err == nil && repo != nil && repo.DefaultBranch != "" {
			config["branch"] = repo.DefaultBranch
		}
	}
}

// resolveTicketConfig resolves ticket identifier from request
// Priority: ticket_identifier > ticket_id
func (h *PodHandler) resolveTicketConfig(c *gin.Context, req *CreatePodRequest, config map[string]interface{}) {
	ctx := c.Request.Context()

	if req.TicketIdentifier != nil && *req.TicketIdentifier != "" {
		config["ticket_identifier"] = *req.TicketIdentifier
	} else if req.TicketID != nil && h.ticketService != nil {
		t, err := h.ticketService.GetTicket(ctx, *req.TicketID)
		if err == nil && t != nil {
			config["ticket_identifier"] = t.Identifier
		}
	}
}

// resolveAgentCredentials resolves agent credentials and injects as env vars
// CredentialProfileID: nil/0 = RunnerHost (no injection), >0 = use profile
func (h *PodHandler) resolveAgentCredentials(c *gin.Context, req *CreatePodRequest, userID int64, config map[string]interface{}) {
	ctx := c.Request.Context()

	if req.AgentTypeID == nil || h.agentService == nil {
		return
	}

	credentials, isRunnerHost, err := h.agentService.GetEffectiveCredentialsForPod(ctx, userID, *req.AgentTypeID, req.CredentialProfileID)
	if err != nil {
		log.Printf("[pods] Failed to get credentials: %v, falling back to RunnerHost mode", err)
		return
	}

	if isRunnerHost || credentials == nil || len(credentials) == 0 {
		return
	}

	// Get agent type slug for env var mapping
	agentType, err := h.agentService.GetAgentType(ctx, *req.AgentTypeID)
	if err != nil || agentType == nil {
		return
	}

	envVars := h.mapCredentialsToEnvVars(agentType.Slug, credentials)
	if len(envVars) == 0 {
		return
	}

	// Merge with existing env_vars
	if existingEnvVars, ok := config["env_vars"].(map[string]interface{}); ok {
		for k, v := range envVars {
			existingEnvVars[k] = v
		}
	} else {
		config["env_vars"] = envVars
	}
	log.Printf("[pods] Injected %d credential env vars for agent %s", len(envVars), agentType.Slug)
}
