package v1

import (
	"context"
	"log"
	"strings"

	"github.com/anthropics/agentmesh/backend/internal/domain/user"
	userService "github.com/anthropics/agentmesh/backend/internal/service/user"
	"github.com/gin-gonic/gin"
)

// UserServiceForPod defines the user service methods needed by PodHandler
// This interface enables easier testing with mock implementations
type UserServiceForPod interface {
	// Git credential methods
	GetDefaultGitCredential(ctx context.Context, userID int64) (*user.GitCredential, error)
	GetDecryptedCredentialToken(ctx context.Context, userID, credentialID int64) (*userService.DecryptedCredential, error)

	// OAuth identity methods
	GetDecryptedTokens(ctx context.Context, userID int64, provider string) (*userService.DecryptedTokens, error)

	// Git connection methods
	GetGitConnectionByProviderAndURL(ctx context.Context, userID int64, providerType, baseURL string) (*user.GitConnection, error)
	GetDecryptedConnectionToken(ctx context.Context, userID, connectionID int64) (*userService.DecryptedTokens, error)
}

// AgentEnvVarMapping defines how credential fields map to environment variables
// for a specific agent type
type AgentEnvVarMapping struct {
	APIKey  string // Environment variable name for API key
	BaseURL string // Environment variable name for base URL
}

// agentEnvVarMappings defines credential field to env var mappings for each agent type
// TODO: Consider moving this to database (agent_types table) for better extensibility
var agentEnvVarMappings = map[string]AgentEnvVarMapping{
	"claude-code": {
		APIKey:  "ANTHROPIC_API_KEY",
		BaseURL: "ANTHROPIC_BASE_URL",
	},
	"codex": {
		APIKey:  "OPENAI_API_KEY",
		BaseURL: "OPENAI_API_BASE",
	},
	"gemini-cli": {
		APIKey: "GEMINI_API_KEY",
	},
	"opencode": {
		APIKey:  "OPENAI_API_KEY",
		BaseURL: "OPENAI_API_BASE",
	},
}

// mapCredentialsToEnvVars converts credential fields to environment variable names
// based on the agent type slug
func (h *PodHandler) mapCredentialsToEnvVars(agentSlug string, credentials map[string]string) map[string]interface{} {
	envVars := make(map[string]interface{})

	// Get mapping for this agent type
	mapping, exists := agentEnvVarMappings[agentSlug]
	if !exists {
		// Default mapping: use uppercase of field name with AGENT_ prefix
		for field, value := range credentials {
			envVars["AGENT_"+strings.ToUpper(field)] = value
		}
		return envVars
	}

	// Apply agent-specific mapping
	for field, value := range credentials {
		switch field {
		case "api_key":
			if mapping.APIKey != "" {
				envVars[mapping.APIKey] = value
			}
		case "base_url":
			if mapping.BaseURL != "" {
				envVars[mapping.BaseURL] = value
			}
		default:
			// Unknown field: use AGENT_ prefix
			envVars["AGENT_"+strings.ToUpper(field)] = value
		}
	}

	return envVars
}

// getUserGitCredential retrieves the default Git credential for the current user
// Implements "权限跟人走" - credentials follow the person, not the organization
//
// Returns:
// - DecryptedCredential with token/ssh key if found
// - nil if using runner_local (Runner will use local Git config)
func (h *PodHandler) getUserGitCredential(c *gin.Context, userID int64) *userService.DecryptedCredential {
	if h.userService == nil {
		return nil
	}

	ctx := c.Request.Context()

	// Get user's default Git credential
	defaultCred, err := h.userService.GetDefaultGitCredential(ctx, userID)
	if err != nil || defaultCred == nil {
		// No default set, use runner_local (return nil)
		return nil
	}

	// If type is runner_local, return nil to let Runner use local config
	if defaultCred.CredentialType == "runner_local" {
		return nil
	}

	// Decrypt and return the credential
	decrypted, err := h.userService.GetDecryptedCredentialToken(ctx, userID, defaultCred.ID)
	if err != nil {
		log.Printf("[pods] Failed to decrypt Git credential: %v", err)
		return nil
	}

	return decrypted
}

// getUserGitToken retrieves the Git access token for the current user
// Implements "权限跟人走" - credentials follow the person, not the organization
//
// Priority:
// 1. OAuth identity matching provider type (for public providers like github.com, gitlab.com)
// 2. PAT connection matching provider type + base URL (for private GitLab instances)
//
// Returns empty string if no credentials found (Runner will use local Git config)
func (h *PodHandler) getUserGitToken(c *gin.Context, userID int64, providerType, providerBaseURL string) string {
	if h.userService == nil {
		return ""
	}

	ctx := c.Request.Context()

	// 1. Try OAuth identity first (for github.com, gitlab.com, gitee.com)
	// OAuth identities only exist for public providers
	if isPublicProvider(providerType, providerBaseURL) {
		tokens, err := h.userService.GetDecryptedTokens(ctx, userID, providerType)
		if err == nil && tokens.AccessToken != "" {
			return tokens.AccessToken
		}
	}

	// 2. Try PAT connections (for private GitLab or additional accounts)
	conn, err := h.userService.GetGitConnectionByProviderAndURL(ctx, userID, providerType, providerBaseURL)
	if err == nil && conn != nil {
		decryptedTokens, err := h.userService.GetDecryptedConnectionToken(ctx, userID, conn.ID)
		if err == nil && decryptedTokens.AccessToken != "" {
			return decryptedTokens.AccessToken
		}
	}

	// No credentials found - Runner will use its local Git configuration
	return ""
}

// isPublicProvider checks if the provider is a public provider (github.com, gitlab.com, gitee.com)
func isPublicProvider(providerType, providerBaseURL string) bool {
	switch providerType {
	case "github":
		return providerBaseURL == "https://github.com" || providerBaseURL == "https://api.github.com"
	case "gitlab":
		return providerBaseURL == "https://gitlab.com"
	case "gitee":
		return providerBaseURL == "https://gitee.com"
	default:
		return false
	}
}
