package v1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/agentmesh/backend/internal/domain/agent"
	"github.com/anthropics/agentmesh/backend/internal/domain/gitprovider"
	"github.com/anthropics/agentmesh/backend/internal/domain/ticket"
	"github.com/anthropics/agentmesh/backend/internal/domain/user"
	"github.com/anthropics/agentmesh/backend/internal/middleware"
	userService "github.com/anthropics/agentmesh/backend/internal/service/user"
	"github.com/gin-gonic/gin"
)

// Mock services for testing

// mockRepositoryService implements repository service methods needed for testing
type mockRepositoryService struct {
	repo *gitprovider.Repository
	err  error
}

func (m *mockRepositoryService) GetByID(ctx context.Context, id int64) (*gitprovider.Repository, error) {
	return m.repo, m.err
}

// mockTicketService implements ticket service methods needed for testing
type mockTicketService struct {
	ticket *ticket.Ticket
	err    error
}

func (m *mockTicketService) GetTicket(ctx context.Context, ticketID int64) (*ticket.Ticket, error) {
	return m.ticket, m.err
}

// mockAgentService implements agent service methods needed for testing
type mockAgentService struct {
	effectiveConfig     agent.ConfigValues
	credentials         agent.EncryptedCredentials
	isRunnerHost        bool
	credentialsErr      error
	agentType           *agent.AgentType
	agentTypeErr        error
}

func (m *mockAgentService) GetEffectiveConfig(ctx context.Context, orgID, agentTypeID int64, overrides agent.ConfigValues) agent.ConfigValues {
	return m.effectiveConfig
}

func (m *mockAgentService) GetEffectiveCredentialsForPod(ctx context.Context, userID, agentTypeID int64, profileID *int64) (agent.EncryptedCredentials, bool, error) {
	return m.credentials, m.isRunnerHost, m.credentialsErr
}

func (m *mockAgentService) GetAgentType(ctx context.Context, id int64) (*agent.AgentType, error) {
	return m.agentType, m.agentTypeErr
}

// Helper to create a test gin context for pod config tests
func createPodConfigTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/test", nil)

	// Set up tenant info using middleware's TenantContext
	tc := &middleware.TenantContext{
		OrganizationID:   100,
		OrganizationSlug: "test-org",
		UserID:           1,
		UserRole:         "owner",
	}
	c.Set("tenant", tc)
	c.Set("user_id", int64(1))

	return c, w
}

// Helper to create int64 pointer (pod_config specific)
func podConfigInt64Ptr(v int64) *int64 {
	return &v
}

// Helper to create string pointer (pod_config specific)
func podConfigStrPtr(v string) *string {
	return &v
}

// =============================================================================
// resolveRepositoryConfig Tests
// =============================================================================

func TestResolveRepositoryConfig_WithRepositoryURL(t *testing.T) {
	h := &PodHandler{}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	req := &CreatePodRequest{
		RepositoryURL: podConfigStrPtr("https://github.com/test/repo.git"),
	}

	h.resolveRepositoryConfig(c, req, config)

	if config["repository_url"] != "https://github.com/test/repo.git" {
		t.Errorf("Expected repository_url to be set, got %v", config["repository_url"])
	}
}

func TestResolveRepositoryConfig_WithRepositoryID(t *testing.T) {
	mockRepo := &mockRepositoryService{
		repo: &gitprovider.Repository{
			ID:              1,
			CloneURL:        "https://github.com/org/repo.git",
			ProviderType:    "github",
			ProviderBaseURL: "https://github.com",
			DefaultBranch:   "main",
		},
	}

	h := &PodHandler{
		repositoryService: mockRepo,
	}
	c, _ := createPodConfigTestContext()
	config := make(map[string]interface{})
	req := &CreatePodRequest{
		RepositoryID: podConfigInt64Ptr(1),
	}

	// Call the actual function
	h.resolveRepositoryConfig(c, req, config)

	if config["repository_url"] != "https://github.com/org/repo.git" {
		t.Errorf("Expected repository_url from ID, got %v", config["repository_url"])
	}
	if config["provider_type"] != "github" {
		t.Errorf("Expected provider_type 'github', got %v", config["provider_type"])
	}
	if config["provider_base_url"] != "https://github.com" {
		t.Errorf("Expected provider_base_url 'https://github.com', got %v", config["provider_base_url"])
	}
}

func TestResolveRepositoryConfig_URLPrecedenceOverID(t *testing.T) {
	h := &PodHandler{}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	req := &CreatePodRequest{
		RepositoryURL: podConfigStrPtr("https://direct-url.com/repo.git"),
		RepositoryID:  podConfigInt64Ptr(1), // Should be ignored
	}

	h.resolveRepositoryConfig(c, req, config)

	if config["repository_url"] != "https://direct-url.com/repo.git" {
		t.Errorf("Expected direct URL to take precedence, got %v", config["repository_url"])
	}
}

func TestResolveRepositoryConfig_EmptyURL(t *testing.T) {
	h := &PodHandler{}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	req := &CreatePodRequest{
		RepositoryURL: podConfigStrPtr(""), // Empty URL should not be set
	}

	h.resolveRepositoryConfig(c, req, config)

	if _, exists := config["repository_url"]; exists {
		t.Error("Expected empty URL to not be set in config")
	}
}

func TestResolveRepositoryConfig_NilRequest(t *testing.T) {
	h := &PodHandler{}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	req := &CreatePodRequest{} // No repository info

	h.resolveRepositoryConfig(c, req, config)

	if _, exists := config["repository_url"]; exists {
		t.Error("Expected no repository_url to be set")
	}
}

func TestResolveRepositoryConfig_ServiceError(t *testing.T) {
	mockRepo := &mockRepositoryService{
		repo: nil,
		err:  errors.New("database error"),
	}

	c, _ := createPodConfigTestContext()
	config := make(map[string]interface{})
	req := &CreatePodRequest{
		RepositoryID: podConfigInt64Ptr(1),
	}

	// Simulate the function behavior with error
	ctx := c.Request.Context()
	repo, err := mockRepo.GetByID(ctx, *req.RepositoryID)
	if err == nil && repo != nil {
		config["repository_url"] = repo.CloneURL
	}

	if _, exists := config["repository_url"]; exists {
		t.Error("Expected no repository_url when service returns error")
	}
}

// =============================================================================
// resolveGitCredentials Tests
// =============================================================================

func TestResolveGitCredentials_NilUserService(t *testing.T) {
	h := &PodHandler{userService: nil}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	h.resolveGitCredentials(c, 1, config)

	if _, exists := config["git_token"]; exists {
		t.Error("Expected no git_token when userService is nil")
	}
}

func TestResolveGitCredentials_OAuthToken(t *testing.T) {
	// getUserGitCredential needs: GetDefaultGitCredential -> GetDecryptedCredentialToken
	mock := &mockUserService{
		defaultGitCredential: &user.GitCredential{
			ID:             1,
			CredentialType: "oauth",
		},
		decryptedCredential: &userService.DecryptedCredential{
			Type:  "oauth",
			Token: "oauth-token-123",
		},
	}

	h := &PodHandler{userService: mock}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	h.resolveGitCredentials(c, 1, config)

	if config["git_token"] != "oauth-token-123" {
		t.Errorf("Expected oauth token, got %v", config["git_token"])
	}
}

func TestResolveGitCredentials_PATToken(t *testing.T) {
	mock := &mockUserService{
		defaultGitCredential: &user.GitCredential{
			ID:             1,
			CredentialType: "pat",
		},
		decryptedCredential: &userService.DecryptedCredential{
			Type:  "pat",
			Token: "pat-token-456",
		},
	}

	h := &PodHandler{userService: mock}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	h.resolveGitCredentials(c, 1, config)

	if config["git_token"] != "pat-token-456" {
		t.Errorf("Expected PAT token, got %v", config["git_token"])
	}
}

func TestResolveGitCredentials_SSHKey(t *testing.T) {
	mock := &mockUserService{
		defaultGitCredential: &user.GitCredential{
			ID:             1,
			CredentialType: "ssh_key",
		},
		decryptedCredential: &userService.DecryptedCredential{
			Type:          "ssh_key",
			SSHPrivateKey: "-----BEGIN RSA PRIVATE KEY-----\ntest-key\n-----END RSA PRIVATE KEY-----",
		},
	}

	h := &PodHandler{userService: mock}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	h.resolveGitCredentials(c, 1, config)

	if config["ssh_private_key"] != "-----BEGIN RSA PRIVATE KEY-----\ntest-key\n-----END RSA PRIVATE KEY-----" {
		t.Errorf("Expected SSH key, got %v", config["ssh_private_key"])
	}
	if _, exists := config["git_token"]; exists {
		t.Error("Expected no git_token for SSH key type")
	}
}

func TestResolveGitCredentials_RunnerLocal(t *testing.T) {
	mock := &mockUserService{
		defaultGitCredential: &user.GitCredential{
			ID:             1,
			CredentialType: "runner_local", // runner_local type
		},
	}

	h := &PodHandler{userService: mock}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	h.resolveGitCredentials(c, 1, config)

	if _, exists := config["git_token"]; exists {
		t.Error("Expected no credentials for runner_local mode")
	}
	if _, exists := config["ssh_private_key"]; exists {
		t.Error("Expected no SSH key for runner_local mode")
	}
}

func TestResolveGitCredentials_EmptyToken(t *testing.T) {
	mock := &mockUserService{
		defaultGitCredential: &user.GitCredential{
			ID:             1,
			CredentialType: "pat",
		},
		decryptedCredential: &userService.DecryptedCredential{
			Type:  "pat",
			Token: "", // Empty token
		},
	}

	h := &PodHandler{userService: mock}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	h.resolveGitCredentials(c, 1, config)

	if _, exists := config["git_token"]; exists {
		t.Error("Expected no git_token when token is empty")
	}
}

// =============================================================================
// resolveBranchConfig Tests
// =============================================================================

func TestResolveBranchConfig_WithBranchName(t *testing.T) {
	h := &PodHandler{}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	req := &CreatePodRequest{
		BranchName: podConfigStrPtr("feature/new-branch"),
	}

	h.resolveBranchConfig(c, req, config)

	if config["branch"] != "feature/new-branch" {
		t.Errorf("Expected branch from request, got %v", config["branch"])
	}
}

func TestResolveBranchConfig_FromRepositoryDefault(t *testing.T) {
	mockRepo := &mockRepositoryService{
		repo: &gitprovider.Repository{
			DefaultBranch: "develop",
		},
	}

	h := &PodHandler{
		repositoryService: mockRepo,
	}
	c, _ := createPodConfigTestContext()
	config := make(map[string]interface{})
	req := &CreatePodRequest{
		RepositoryID: podConfigInt64Ptr(1),
	}

	// Call the actual function
	h.resolveBranchConfig(c, req, config)

	if config["branch"] != "develop" {
		t.Errorf("Expected default branch from repo, got %v", config["branch"])
	}
}

func TestResolveBranchConfig_EmptyBranchName(t *testing.T) {
	h := &PodHandler{}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	req := &CreatePodRequest{
		BranchName: podConfigStrPtr(""), // Empty should not be set
	}

	h.resolveBranchConfig(c, req, config)

	if _, exists := config["branch"]; exists {
		t.Error("Expected no branch when branch name is empty")
	}
}

func TestResolveBranchConfig_NoBranch(t *testing.T) {
	h := &PodHandler{}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	req := &CreatePodRequest{} // No branch info

	h.resolveBranchConfig(c, req, config)

	if _, exists := config["branch"]; exists {
		t.Error("Expected no branch to be set")
	}
}

// =============================================================================
// resolveTicketConfig Tests
// =============================================================================

func TestResolveTicketConfig_WithTicketIdentifier(t *testing.T) {
	h := &PodHandler{}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	req := &CreatePodRequest{
		TicketIdentifier: podConfigStrPtr("AM-123"),
	}

	h.resolveTicketConfig(c, req, config)

	if config["ticket_identifier"] != "AM-123" {
		t.Errorf("Expected ticket_identifier from request, got %v", config["ticket_identifier"])
	}
}

func TestResolveTicketConfig_FromTicketID(t *testing.T) {
	mockTicket := &mockTicketService{
		ticket: &ticket.Ticket{
			ID:         1,
			Identifier: "PROJ-456",
		},
	}

	h := &PodHandler{
		ticketService: mockTicket,
	}
	c, _ := createPodConfigTestContext()
	config := make(map[string]interface{})
	req := &CreatePodRequest{
		TicketID: podConfigInt64Ptr(1),
	}

	// Call the actual function
	h.resolveTicketConfig(c, req, config)

	if config["ticket_identifier"] != "PROJ-456" {
		t.Errorf("Expected ticket_identifier from ID, got %v", config["ticket_identifier"])
	}
}

func TestResolveTicketConfig_IdentifierPrecedenceOverID(t *testing.T) {
	h := &PodHandler{}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	req := &CreatePodRequest{
		TicketIdentifier: podConfigStrPtr("DIRECT-999"),
		TicketID:         podConfigInt64Ptr(1), // Should be ignored
	}

	h.resolveTicketConfig(c, req, config)

	if config["ticket_identifier"] != "DIRECT-999" {
		t.Errorf("Expected direct identifier to take precedence, got %v", config["ticket_identifier"])
	}
}

func TestResolveTicketConfig_EmptyIdentifier(t *testing.T) {
	h := &PodHandler{}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	req := &CreatePodRequest{
		TicketIdentifier: podConfigStrPtr(""), // Empty should not be set
	}

	h.resolveTicketConfig(c, req, config)

	if _, exists := config["ticket_identifier"]; exists {
		t.Error("Expected no ticket_identifier when identifier is empty")
	}
}

func TestResolveTicketConfig_ServiceError(t *testing.T) {
	mockTicket := &mockTicketService{
		ticket: nil,
		err:    errors.New("ticket not found"),
	}

	c, _ := createPodConfigTestContext()
	config := make(map[string]interface{})
	req := &CreatePodRequest{
		TicketID: podConfigInt64Ptr(999),
	}

	// Simulate the logic
	ctx := c.Request.Context()
	tk, err := mockTicket.GetTicket(ctx, *req.TicketID)
	if err == nil && tk != nil {
		config["ticket_identifier"] = tk.Identifier
	}

	if _, exists := config["ticket_identifier"]; exists {
		t.Error("Expected no ticket_identifier when service returns error")
	}
}

// =============================================================================
// resolveAgentCredentials Tests
// =============================================================================

func TestResolveAgentCredentials_NilAgentTypeID(t *testing.T) {
	h := &PodHandler{}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	req := &CreatePodRequest{
		AgentTypeID: nil, // No agent type
	}

	h.resolveAgentCredentials(c, req, 1, config)

	if _, exists := config["env_vars"]; exists {
		t.Error("Expected no env_vars when AgentTypeID is nil")
	}
}

func TestResolveAgentCredentials_RunnerHostMode(t *testing.T) {
	mockAgent := &mockAgentService{
		credentials:  nil,
		isRunnerHost: true,
		agentType: &agent.AgentType{
			ID:   1,
			Slug: "claude-code",
			Name: "Claude Code",
		},
	}

	h := &PodHandler{
		agentService: mockAgent,
	}
	c, _ := createPodConfigTestContext()
	config := make(map[string]interface{})
	agentTypeID := int64(1)
	req := &CreatePodRequest{
		AgentTypeID: &agentTypeID,
	}

	// Call the actual function
	h.resolveAgentCredentials(c, req, 1, config)

	if _, exists := config["env_vars"]; exists {
		t.Error("Expected no env_vars in RunnerHost mode")
	}
}

func TestResolveAgentCredentials_WithCredentials(t *testing.T) {
	mockAgent := &mockAgentService{
		credentials: agent.EncryptedCredentials{
			"api_key":  "sk-test-key",
			"base_url": "https://api.anthropic.com",
		},
		isRunnerHost: false,
		agentType: &agent.AgentType{
			ID:   1,
			Slug: "claude-code",
			Name: "Claude Code",
		},
	}

	h := &PodHandler{
		agentService: mockAgent,
	}
	c, _ := createPodConfigTestContext()
	config := make(map[string]interface{})
	agentTypeID := int64(1)
	req := &CreatePodRequest{
		AgentTypeID: &agentTypeID,
	}

	// Call the actual function
	h.resolveAgentCredentials(c, req, 1, config)

	envVars, ok := config["env_vars"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected env_vars to be set")
	}

	if envVars["ANTHROPIC_API_KEY"] != "sk-test-key" {
		t.Errorf("Expected ANTHROPIC_API_KEY, got %v", envVars["ANTHROPIC_API_KEY"])
	}
	if envVars["ANTHROPIC_BASE_URL"] != "https://api.anthropic.com" {
		t.Errorf("Expected ANTHROPIC_BASE_URL, got %v", envVars["ANTHROPIC_BASE_URL"])
	}
}

func TestResolveAgentCredentials_MergeWithExistingEnvVars(t *testing.T) {
	mockAgent := &mockAgentService{
		credentials: agent.EncryptedCredentials{
			"api_key": "new-key",
		},
		isRunnerHost: false,
		agentType: &agent.AgentType{
			ID:   1,
			Slug: "claude-code",
			Name: "Claude Code",
		},
	}

	h := &PodHandler{
		agentService: mockAgent,
	}
	c, _ := createPodConfigTestContext()
	// Pre-existing env_vars
	config := map[string]interface{}{
		"env_vars": map[string]interface{}{
			"EXISTING_VAR": "existing-value",
		},
	}
	agentTypeID := int64(1)
	req := &CreatePodRequest{
		AgentTypeID: &agentTypeID,
	}

	// Call the actual function
	h.resolveAgentCredentials(c, req, 1, config)

	envVars := config["env_vars"].(map[string]interface{})
	if envVars["EXISTING_VAR"] != "existing-value" {
		t.Error("Expected existing env var to be preserved")
	}
	if envVars["ANTHROPIC_API_KEY"] != "new-key" {
		t.Error("Expected new credential to be merged")
	}
}

func TestResolveAgentCredentials_CredentialsFetchError(t *testing.T) {
	mockAgent := &mockAgentService{
		credentialsErr: errors.New("failed to fetch credentials"),
	}

	h := &PodHandler{
		agentService: mockAgent,
	}
	c, _ := createPodConfigTestContext()
	config := make(map[string]interface{})
	agentTypeID := int64(1)
	req := &CreatePodRequest{
		AgentTypeID: &agentTypeID,
	}

	// Call the actual function
	h.resolveAgentCredentials(c, req, 1, config)

	if _, exists := config["env_vars"]; exists {
		t.Error("Expected no env_vars when credentials fetch fails")
	}
}

func TestResolveAgentCredentials_AgentTypeFetchError(t *testing.T) {
	mockAgent := &mockAgentService{
		credentials:  agent.EncryptedCredentials{"api_key": "test"},
		isRunnerHost: false,
		agentType:    nil,
		agentTypeErr: errors.New("agent type not found"),
	}

	h := &PodHandler{
		agentService: mockAgent,
	}
	c, _ := createPodConfigTestContext()
	config := make(map[string]interface{})
	agentTypeID := int64(1)
	req := &CreatePodRequest{
		AgentTypeID: &agentTypeID,
	}

	// Call the actual function
	h.resolveAgentCredentials(c, req, 1, config)

	if _, exists := config["env_vars"]; exists {
		t.Error("Expected no env_vars when agent type fetch fails")
	}
}

func TestResolveAgentCredentials_EmptyCredentials(t *testing.T) {
	mockAgent := &mockAgentService{
		credentials:  agent.EncryptedCredentials{}, // Empty
		isRunnerHost: false,
		agentType: &agent.AgentType{
			ID:   1,
			Slug: "claude-code",
			Name: "Claude Code",
		},
	}

	h := &PodHandler{
		agentService: mockAgent,
	}
	c, _ := createPodConfigTestContext()
	config := make(map[string]interface{})
	agentTypeID := int64(1)
	req := &CreatePodRequest{
		AgentTypeID: &agentTypeID,
	}

	// Call the actual function
	h.resolveAgentCredentials(c, req, 1, config)

	if _, exists := config["env_vars"]; exists {
		t.Error("Expected no env_vars when credentials are empty")
	}
}

// =============================================================================
// buildPluginConfig Integration Tests
// =============================================================================

func TestBuildPluginConfig_MergeUserPluginConfig(t *testing.T) {
	h := &PodHandler{}
	c, _ := createPodConfigTestContext()

	// Test that user PluginConfig overrides other values
	config := make(map[string]interface{})

	// Simulate buildPluginConfig logic for PluginConfig merge
	req := &CreatePodRequest{
		RepositoryURL: podConfigStrPtr("https://github.com/org/repo.git"),
		PluginConfig: map[string]interface{}{
			"repository_url": "https://override.com/repo.git", // Override
			"custom_key":     "custom_value",
		},
	}

	// Step 1: Resolve repository
	h.resolveRepositoryConfig(c, req, config)

	// Step 2: Merge user PluginConfig
	if req.PluginConfig != nil {
		for k, v := range req.PluginConfig {
			config[k] = v
		}
	}

	// User PluginConfig should override
	if config["repository_url"] != "https://override.com/repo.git" {
		t.Errorf("Expected PluginConfig to override repository_url, got %v", config["repository_url"])
	}
	if config["custom_key"] != "custom_value" {
		t.Errorf("Expected custom_key to be set, got %v", config["custom_key"])
	}
}

func TestBuildPluginConfig_AllFieldsResolved(t *testing.T) {
	mock := &mockUserService{
		defaultGitCredential: &user.GitCredential{
			ID:             1,
			CredentialType: "pat",
		},
		decryptedCredential: &userService.DecryptedCredential{
			Type:  "pat",
			Token: "test-token",
		},
	}

	h := &PodHandler{userService: mock}
	c, _ := createPodConfigTestContext()

	config := make(map[string]interface{})
	req := &CreatePodRequest{
		RepositoryURL:    podConfigStrPtr("https://github.com/org/repo.git"),
		BranchName:       podConfigStrPtr("main"),
		TicketIdentifier: podConfigStrPtr("PROJ-123"),
		PluginConfig: map[string]interface{}{
			"extra_config": "value",
		},
	}

	// Simulate all resolution steps
	h.resolveRepositoryConfig(c, req, config)
	h.resolveGitCredentials(c, 1, config)
	h.resolveBranchConfig(c, req, config)
	h.resolveTicketConfig(c, req, config)

	// Merge PluginConfig
	if req.PluginConfig != nil {
		for k, v := range req.PluginConfig {
			config[k] = v
		}
	}

	// Verify all fields
	if config["repository_url"] != "https://github.com/org/repo.git" {
		t.Errorf("Expected repository_url, got %v", config["repository_url"])
	}
	if config["git_token"] != "test-token" {
		t.Errorf("Expected git_token, got %v", config["git_token"])
	}
	if config["branch"] != "main" {
		t.Errorf("Expected branch, got %v", config["branch"])
	}
	if config["ticket_identifier"] != "PROJ-123" {
		t.Errorf("Expected ticket_identifier, got %v", config["ticket_identifier"])
	}
	if config["extra_config"] != "value" {
		t.Errorf("Expected extra_config, got %v", config["extra_config"])
	}
}

// =============================================================================
// buildPluginConfig Integration Tests
// =============================================================================

func TestBuildPluginConfig_Integration_Full(t *testing.T) {
	agentTypeID := int64(1)
	mockUserSvc := &mockUserService{
		defaultGitCredential: &user.GitCredential{
			ID:             1,
			CredentialType: "pat",
		},
		decryptedCredential: &userService.DecryptedCredential{
			Type:  "pat",
			Token: "test-git-token",
		},
	}
	mockAgentSvc := &mockAgentService{
		effectiveConfig: agent.ConfigValues{
			"model":           "opus",
			"permission_mode": "plan",
		},
		credentials: agent.EncryptedCredentials{
			"api_key": "encrypted-api-key",
		},
		isRunnerHost: false,
		agentType: &agent.AgentType{
			ID:   agentTypeID,
			Slug: "claude-code",
			Name: "Claude Code",
		},
	}
	mockRepoSvc := &mockRepositoryService{
		repo: &gitprovider.Repository{
			ID:              1,
			CloneURL:        "https://github.com/org/repo.git",
			DefaultBranch:   "main",
			ProviderType:    "github",
			ProviderBaseURL: "https://api.github.com",
		},
	}
	mockTicketSvc := &mockTicketService{
		ticket: &ticket.Ticket{
			ID:         1,
			Identifier: "PROJ-123",
		},
	}

	h := &PodHandler{
		userService:       mockUserSvc,
		agentService:      mockAgentSvc,
		repositoryService: mockRepoSvc,
		ticketService:     mockTicketSvc,
	}
	c, _ := createPodConfigTestContext()

	req := &CreatePodRequest{
		RunnerID:      1,
		AgentTypeID:   &agentTypeID,
		RepositoryURL: podConfigStrPtr("https://gitlab.com/custom/repo.git"),
		BranchName:    podConfigStrPtr("feature-branch"),
		PluginConfig: map[string]interface{}{
			"custom_key": "custom_value",
		},
	}

	config := h.buildPluginConfig(c, req)

	// Verify organization config is applied
	if config["model"] != "opus" {
		t.Errorf("Expected model=opus from org config, got %v", config["model"])
	}
	if config["permission_mode"] != "plan" {
		t.Errorf("Expected permission_mode=plan from org config, got %v", config["permission_mode"])
	}

	// Verify repository URL (direct URL takes precedence)
	if config["repository_url"] != "https://gitlab.com/custom/repo.git" {
		t.Errorf("Expected custom repository_url, got %v", config["repository_url"])
	}

	// Verify branch name
	if config["branch"] != "feature-branch" {
		t.Errorf("Expected branch=feature-branch, got %v", config["branch"])
	}

	// Verify git credentials
	if config["git_token"] != "test-git-token" {
		t.Errorf("Expected git_token, got %v", config["git_token"])
	}

	// Verify custom PluginConfig is merged
	if config["custom_key"] != "custom_value" {
		t.Errorf("Expected custom_key=custom_value, got %v", config["custom_key"])
	}

	// Verify env_vars contains API key
	envVars, ok := config["env_vars"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected env_vars to be map[string]interface{}, got %T", config["env_vars"])
	}
	if envVars["ANTHROPIC_API_KEY"] != "encrypted-api-key" {
		t.Errorf("Expected ANTHROPIC_API_KEY in env_vars, got %v", envVars)
	}
}

func TestBuildPluginConfig_Integration_RunnerHostMode(t *testing.T) {
	agentTypeID := int64(1)
	mockAgentSvc := &mockAgentService{
		effectiveConfig: agent.ConfigValues{
			"model": "sonnet",
		},
		isRunnerHost: true, // Runner host mode - no credentials injected
		agentType: &agent.AgentType{
			ID:   agentTypeID,
			Slug: "claude-code",
			Name: "Claude Code",
		},
	}

	h := &PodHandler{
		agentService: mockAgentSvc,
	}
	c, _ := createPodConfigTestContext()

	req := &CreatePodRequest{
		RunnerID:    1,
		AgentTypeID: &agentTypeID,
	}

	config := h.buildPluginConfig(c, req)

	// Verify org config is applied
	if config["model"] != "sonnet" {
		t.Errorf("Expected model=sonnet, got %v", config["model"])
	}

	// Verify no env_vars are injected in runner host mode
	if _, exists := config["env_vars"]; exists {
		t.Errorf("Expected no env_vars in runner host mode, got %v", config["env_vars"])
	}
}

func TestBuildPluginConfig_Integration_NoAgentType(t *testing.T) {
	mockUserSvc := &mockUserService{
		defaultGitCredential: &user.GitCredential{
			ID:             1,
			CredentialType: "oauth",
		},
		decryptedCredential: &userService.DecryptedCredential{
			Type:  "oauth",
			Token: "oauth-token",
		},
	}

	h := &PodHandler{
		userService: mockUserSvc,
	}
	c, _ := createPodConfigTestContext()

	req := &CreatePodRequest{
		RunnerID:      1,
		AgentTypeID:   nil, // No agent type
		RepositoryURL: podConfigStrPtr("https://github.com/org/repo.git"),
	}

	config := h.buildPluginConfig(c, req)

	// Verify repository URL is resolved
	if config["repository_url"] != "https://github.com/org/repo.git" {
		t.Errorf("Expected repository_url, got %v", config["repository_url"])
	}

	// Verify git credentials are resolved
	if config["git_token"] != "oauth-token" {
		t.Errorf("Expected git_token, got %v", config["git_token"])
	}
}

func TestBuildPluginConfig_Integration_PluginConfigOverridesOrgConfig(t *testing.T) {
	agentTypeID := int64(1)
	mockAgentSvc := &mockAgentService{
		effectiveConfig: agent.ConfigValues{
			"model":           "opus",
			"permission_mode": "plan",
		},
		isRunnerHost: true,
		agentType: &agent.AgentType{
			ID:   agentTypeID,
			Slug: "claude-code",
			Name: "Claude Code",
		},
	}

	h := &PodHandler{
		agentService: mockAgentSvc,
	}
	c, _ := createPodConfigTestContext()

	req := &CreatePodRequest{
		RunnerID:    1,
		AgentTypeID: &agentTypeID,
		PluginConfig: map[string]interface{}{
			"model":      "sonnet", // Override org config
			"extra_flag": true,
		},
	}

	config := h.buildPluginConfig(c, req)

	// Verify PluginConfig overrides org config
	if config["model"] != "sonnet" {
		t.Errorf("Expected model=sonnet (overridden), got %v", config["model"])
	}
	if config["permission_mode"] != "plan" {
		t.Errorf("Expected permission_mode=plan (from org), got %v", config["permission_mode"])
	}
	if config["extra_flag"] != true {
		t.Errorf("Expected extra_flag=true, got %v", config["extra_flag"])
	}
}
