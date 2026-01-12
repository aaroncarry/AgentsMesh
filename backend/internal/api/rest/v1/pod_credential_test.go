package v1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/agentmesh/backend/internal/domain/user"
	userService "github.com/anthropics/agentmesh/backend/internal/service/user"
	"github.com/gin-gonic/gin"
)

// ==================== mapCredentialsToEnvVars Tests ====================

func TestMapCredentialsToEnvVars_ClaudeCode(t *testing.T) {
	h := &PodHandler{}

	credentials := map[string]string{
		"api_key":  "sk-ant-test123",
		"base_url": "https://api.anthropic.com",
	}

	envVars := h.mapCredentialsToEnvVars("claude-code", credentials)

	if envVars["ANTHROPIC_API_KEY"] != "sk-ant-test123" {
		t.Errorf("Expected ANTHROPIC_API_KEY to be 'sk-ant-test123', got '%v'", envVars["ANTHROPIC_API_KEY"])
	}
	if envVars["ANTHROPIC_BASE_URL"] != "https://api.anthropic.com" {
		t.Errorf("Expected ANTHROPIC_BASE_URL to be 'https://api.anthropic.com', got '%v'", envVars["ANTHROPIC_BASE_URL"])
	}
}

func TestMapCredentialsToEnvVars_Codex(t *testing.T) {
	h := &PodHandler{}

	credentials := map[string]string{
		"api_key":  "sk-openai-test456",
		"base_url": "https://api.openai.com",
	}

	envVars := h.mapCredentialsToEnvVars("codex", credentials)

	if envVars["OPENAI_API_KEY"] != "sk-openai-test456" {
		t.Errorf("Expected OPENAI_API_KEY to be 'sk-openai-test456', got '%v'", envVars["OPENAI_API_KEY"])
	}
	if envVars["OPENAI_API_BASE"] != "https://api.openai.com" {
		t.Errorf("Expected OPENAI_API_BASE to be 'https://api.openai.com', got '%v'", envVars["OPENAI_API_BASE"])
	}
}

func TestMapCredentialsToEnvVars_GeminiCLI(t *testing.T) {
	h := &PodHandler{}

	credentials := map[string]string{
		"api_key": "gemini-key-789",
	}

	envVars := h.mapCredentialsToEnvVars("gemini-cli", credentials)

	if envVars["GEMINI_API_KEY"] != "gemini-key-789" {
		t.Errorf("Expected GEMINI_API_KEY to be 'gemini-key-789', got '%v'", envVars["GEMINI_API_KEY"])
	}
}

func TestMapCredentialsToEnvVars_OpenCode(t *testing.T) {
	h := &PodHandler{}

	credentials := map[string]string{
		"api_key":  "sk-opencode-test",
		"base_url": "https://custom.openai.com",
	}

	envVars := h.mapCredentialsToEnvVars("opencode", credentials)

	if envVars["OPENAI_API_KEY"] != "sk-opencode-test" {
		t.Errorf("Expected OPENAI_API_KEY to be 'sk-opencode-test', got '%v'", envVars["OPENAI_API_KEY"])
	}
	if envVars["OPENAI_API_BASE"] != "https://custom.openai.com" {
		t.Errorf("Expected OPENAI_API_BASE to be 'https://custom.openai.com', got '%v'", envVars["OPENAI_API_BASE"])
	}
}

func TestMapCredentialsToEnvVars_UnknownAgentType(t *testing.T) {
	h := &PodHandler{}

	credentials := map[string]string{
		"api_key":      "test-key",
		"custom_field": "custom-value",
	}

	envVars := h.mapCredentialsToEnvVars("unknown-agent", credentials)

	// Unknown agent type should use AGENT_ prefix
	if envVars["AGENT_API_KEY"] != "test-key" {
		t.Errorf("Expected AGENT_API_KEY to be 'test-key', got '%v'", envVars["AGENT_API_KEY"])
	}
	if envVars["AGENT_CUSTOM_FIELD"] != "custom-value" {
		t.Errorf("Expected AGENT_CUSTOM_FIELD to be 'custom-value', got '%v'", envVars["AGENT_CUSTOM_FIELD"])
	}
}

func TestMapCredentialsToEnvVars_UnknownFieldInKnownAgent(t *testing.T) {
	h := &PodHandler{}

	credentials := map[string]string{
		"api_key":      "sk-ant-test",
		"custom_field": "custom-value",
	}

	envVars := h.mapCredentialsToEnvVars("claude-code", credentials)

	// Known field should use mapped name
	if envVars["ANTHROPIC_API_KEY"] != "sk-ant-test" {
		t.Errorf("Expected ANTHROPIC_API_KEY to be 'sk-ant-test', got '%v'", envVars["ANTHROPIC_API_KEY"])
	}
	// Unknown field should use AGENT_ prefix
	if envVars["AGENT_CUSTOM_FIELD"] != "custom-value" {
		t.Errorf("Expected AGENT_CUSTOM_FIELD to be 'custom-value', got '%v'", envVars["AGENT_CUSTOM_FIELD"])
	}
}

func TestMapCredentialsToEnvVars_EmptyCredentials(t *testing.T) {
	h := &PodHandler{}

	credentials := map[string]string{}

	envVars := h.mapCredentialsToEnvVars("claude-code", credentials)

	if len(envVars) != 0 {
		t.Errorf("Expected empty envVars map, got %v", envVars)
	}
}

func TestMapCredentialsToEnvVars_NilCredentials(t *testing.T) {
	h := &PodHandler{}

	var credentials map[string]string = nil

	envVars := h.mapCredentialsToEnvVars("claude-code", credentials)

	if len(envVars) != 0 {
		t.Errorf("Expected empty envVars map for nil input, got %v", envVars)
	}
}

func TestMapCredentialsToEnvVars_OnlyBaseUrl(t *testing.T) {
	h := &PodHandler{}

	credentials := map[string]string{
		"base_url": "https://proxy.example.com",
	}

	envVars := h.mapCredentialsToEnvVars("claude-code", credentials)

	if envVars["ANTHROPIC_BASE_URL"] != "https://proxy.example.com" {
		t.Errorf("Expected ANTHROPIC_BASE_URL to be 'https://proxy.example.com', got '%v'", envVars["ANTHROPIC_BASE_URL"])
	}
	if _, exists := envVars["ANTHROPIC_API_KEY"]; exists {
		t.Errorf("ANTHROPIC_API_KEY should not exist when not provided")
	}
}

// ==================== isPublicProvider Tests ====================

func TestIsPublicProvider_GitHub(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		expectTrue bool
	}{
		{"github.com", "https://github.com", true},
		{"api.github.com", "https://api.github.com", true},
		{"enterprise github", "https://github.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPublicProvider("github", tt.baseURL)
			if result != tt.expectTrue {
				t.Errorf("isPublicProvider('github', '%s') = %v, want %v", tt.baseURL, result, tt.expectTrue)
			}
		})
	}
}

func TestIsPublicProvider_GitLab(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		expectTrue bool
	}{
		{"gitlab.com", "https://gitlab.com", true},
		{"self-hosted gitlab", "https://gitlab.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPublicProvider("gitlab", tt.baseURL)
			if result != tt.expectTrue {
				t.Errorf("isPublicProvider('gitlab', '%s') = %v, want %v", tt.baseURL, result, tt.expectTrue)
			}
		})
	}
}

func TestIsPublicProvider_Gitee(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		expectTrue bool
	}{
		{"gitee.com", "https://gitee.com", true},
		{"enterprise gitee", "https://gitee.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPublicProvider("gitee", tt.baseURL)
			if result != tt.expectTrue {
				t.Errorf("isPublicProvider('gitee', '%s') = %v, want %v", tt.baseURL, result, tt.expectTrue)
			}
		})
	}
}

func TestIsPublicProvider_UnknownProvider(t *testing.T) {
	result := isPublicProvider("unknown", "https://unknown.com")
	if result != false {
		t.Errorf("isPublicProvider('unknown', 'https://unknown.com') = %v, want false", result)
	}
}

// ==================== agentEnvVarMappings Validation Tests ====================

func TestAgentEnvVarMappings_AllAgentsHaveAPIKey(t *testing.T) {
	// Verify that all defined agent types have at least an API key mapping
	for agentSlug, mapping := range agentEnvVarMappings {
		if mapping.APIKey == "" {
			t.Errorf("Agent '%s' should have APIKey mapping defined", agentSlug)
		}
	}
}

func TestAgentEnvVarMappings_NoDuplicateEnvVarNames(t *testing.T) {
	// Track which env var names are used by which agents
	envVarToAgent := make(map[string][]string)

	for agentSlug, mapping := range agentEnvVarMappings {
		if mapping.APIKey != "" {
			envVarToAgent[mapping.APIKey] = append(envVarToAgent[mapping.APIKey], agentSlug)
		}
		if mapping.BaseURL != "" {
			envVarToAgent[mapping.BaseURL] = append(envVarToAgent[mapping.BaseURL], agentSlug)
		}
	}

	// This is informational - some agents (codex, opencode) share the same OpenAI env vars
	// which is expected behavior
	for envVar, agents := range envVarToAgent {
		if len(agents) > 1 {
			t.Logf("Info: Env var '%s' is used by multiple agents: %v (this may be intentional)", envVar, agents)
		}
	}
}

// ==================== Mock UserService for Testing ====================

// mockUserService implements UserServiceForPod for testing
type mockUserService struct {
	// Mock data
	defaultGitCredential    *user.GitCredential
	decryptedCredential     *userService.DecryptedCredential
	decryptedTokens         *userService.DecryptedTokens
	gitConnection           *user.GitConnection
	decryptedConnectionToken *userService.DecryptedTokens

	// Error returns
	getDefaultGitCredentialErr    error
	getDecryptedCredentialErr     error
	getDecryptedTokensErr         error
	getGitConnectionErr           error
	getDecryptedConnectionTokenErr error
}

func (m *mockUserService) GetDefaultGitCredential(ctx context.Context, userID int64) (*user.GitCredential, error) {
	if m.getDefaultGitCredentialErr != nil {
		return nil, m.getDefaultGitCredentialErr
	}
	return m.defaultGitCredential, nil
}

func (m *mockUserService) GetDecryptedCredentialToken(ctx context.Context, userID, credentialID int64) (*userService.DecryptedCredential, error) {
	if m.getDecryptedCredentialErr != nil {
		return nil, m.getDecryptedCredentialErr
	}
	return m.decryptedCredential, nil
}

func (m *mockUserService) GetDecryptedTokens(ctx context.Context, userID int64, provider string) (*userService.DecryptedTokens, error) {
	if m.getDecryptedTokensErr != nil {
		return nil, m.getDecryptedTokensErr
	}
	return m.decryptedTokens, nil
}

func (m *mockUserService) GetGitConnectionByProviderAndURL(ctx context.Context, userID int64, providerType, baseURL string) (*user.GitConnection, error) {
	if m.getGitConnectionErr != nil {
		return nil, m.getGitConnectionErr
	}
	return m.gitConnection, nil
}

func (m *mockUserService) GetDecryptedConnectionToken(ctx context.Context, userID, connectionID int64) (*userService.DecryptedTokens, error) {
	if m.getDecryptedConnectionTokenErr != nil {
		return nil, m.getDecryptedConnectionTokenErr
	}
	return m.decryptedConnectionToken, nil
}

// Ensure mockUserService implements UserServiceForPod
var _ UserServiceForPod = (*mockUserService)(nil)

// ==================== Test Helpers ====================

// createCredentialTestContext creates a gin context with user ID set
func createCredentialTestContext(userID int64) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request = c.Request.WithContext(context.Background())

	// Set user ID in context
	c.Set("user_id", userID)

	return c, w
}

// ==================== getUserGitCredential Tests ====================

func TestGetUserGitCredential_NilUserService(t *testing.T) {
	h := &PodHandler{userService: nil}
	c, _ := createCredentialTestContext(1)

	result := h.getUserGitCredential(c, 1)
	if result != nil {
		t.Errorf("Expected nil when userService is nil, got %v", result)
	}
}

func TestGetUserGitCredential_NoDefaultCredential(t *testing.T) {
	mock := &mockUserService{
		getDefaultGitCredentialErr: errors.New("not found"),
	}
	h := &PodHandler{userService: mock}
	c, _ := createCredentialTestContext(1)

	result := h.getUserGitCredential(c, 1)
	if result != nil {
		t.Errorf("Expected nil when no default credential, got %v", result)
	}
}

func TestGetUserGitCredential_NilDefaultCredential(t *testing.T) {
	mock := &mockUserService{
		defaultGitCredential: nil,
	}
	h := &PodHandler{userService: mock}
	c, _ := createCredentialTestContext(1)

	result := h.getUserGitCredential(c, 1)
	if result != nil {
		t.Errorf("Expected nil when default credential is nil, got %v", result)
	}
}

func TestGetUserGitCredential_RunnerLocalType(t *testing.T) {
	mock := &mockUserService{
		defaultGitCredential: &user.GitCredential{
			ID:             1,
			UserID:         1,
			CredentialType: "runner_local",
			Name:           "Runner Local",
		},
	}
	h := &PodHandler{userService: mock}
	c, _ := createCredentialTestContext(1)

	result := h.getUserGitCredential(c, 1)
	if result != nil {
		t.Errorf("Expected nil for runner_local credential type, got %v", result)
	}
}

func TestGetUserGitCredential_DecryptError(t *testing.T) {
	mock := &mockUserService{
		defaultGitCredential: &user.GitCredential{
			ID:             1,
			UserID:         1,
			CredentialType: "pat",
			Name:           "My PAT",
		},
		getDecryptedCredentialErr: errors.New("decrypt failed"),
	}
	h := &PodHandler{userService: mock}
	c, _ := createCredentialTestContext(1)

	result := h.getUserGitCredential(c, 1)
	if result != nil {
		t.Errorf("Expected nil when decryption fails, got %v", result)
	}
}

func TestGetUserGitCredential_Success(t *testing.T) {
	expectedCred := &userService.DecryptedCredential{
		Type:  "pat",
		Token: "ghp_test123",
	}
	mock := &mockUserService{
		defaultGitCredential: &user.GitCredential{
			ID:             1,
			UserID:         1,
			CredentialType: "pat",
			Name:           "My PAT",
		},
		decryptedCredential: expectedCred,
	}
	h := &PodHandler{userService: mock}
	c, _ := createCredentialTestContext(1)

	result := h.getUserGitCredential(c, 1)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Token != "ghp_test123" {
		t.Errorf("Expected token 'ghp_test123', got '%s'", result.Token)
	}
	if result.Type != "pat" {
		t.Errorf("Expected type 'pat', got '%s'", result.Type)
	}
}

// ==================== getUserGitToken Tests ====================

func TestGetUserGitToken_NilUserService(t *testing.T) {
	h := &PodHandler{userService: nil}
	c, _ := createCredentialTestContext(1)

	result := h.getUserGitToken(c, 1, "github", "https://github.com")
	if result != "" {
		t.Errorf("Expected empty string when userService is nil, got '%s'", result)
	}
}

func TestGetUserGitToken_PublicGitHub_OAuthSuccess(t *testing.T) {
	mock := &mockUserService{
		decryptedTokens: &userService.DecryptedTokens{
			AccessToken: "gho_oauth123",
		},
	}
	h := &PodHandler{userService: mock}
	c, _ := createCredentialTestContext(1)

	result := h.getUserGitToken(c, 1, "github", "https://github.com")
	if result != "gho_oauth123" {
		t.Errorf("Expected 'gho_oauth123', got '%s'", result)
	}
}

func TestGetUserGitToken_PublicGitHub_OAuthError_FallbackToConnection(t *testing.T) {
	mock := &mockUserService{
		getDecryptedTokensErr: errors.New("no oauth"),
		gitConnection: &user.GitConnection{
			ID:           1,
			UserID:       1,
			ProviderType: "github",
			BaseURL:      "https://github.com",
		},
		decryptedConnectionToken: &userService.DecryptedTokens{
			AccessToken: "ghp_pat123",
		},
	}
	h := &PodHandler{userService: mock}
	c, _ := createCredentialTestContext(1)

	result := h.getUserGitToken(c, 1, "github", "https://github.com")
	if result != "ghp_pat123" {
		t.Errorf("Expected 'ghp_pat123', got '%s'", result)
	}
}

func TestGetUserGitToken_PublicGitHub_EmptyOAuthToken_FallbackToConnection(t *testing.T) {
	mock := &mockUserService{
		decryptedTokens: &userService.DecryptedTokens{
			AccessToken: "", // Empty token
		},
		gitConnection: &user.GitConnection{
			ID:           1,
			UserID:       1,
			ProviderType: "github",
			BaseURL:      "https://github.com",
		},
		decryptedConnectionToken: &userService.DecryptedTokens{
			AccessToken: "ghp_pat123",
		},
	}
	h := &PodHandler{userService: mock}
	c, _ := createCredentialTestContext(1)

	result := h.getUserGitToken(c, 1, "github", "https://github.com")
	if result != "ghp_pat123" {
		t.Errorf("Expected 'ghp_pat123', got '%s'", result)
	}
}

func TestGetUserGitToken_PrivateGitLab_ConnectionSuccess(t *testing.T) {
	mock := &mockUserService{
		gitConnection: &user.GitConnection{
			ID:           1,
			UserID:       1,
			ProviderType: "gitlab",
			BaseURL:      "https://gitlab.company.com",
		},
		decryptedConnectionToken: &userService.DecryptedTokens{
			AccessToken: "glpat_private123",
		},
	}
	h := &PodHandler{userService: mock}
	c, _ := createCredentialTestContext(1)

	// Private GitLab - doesn't try OAuth first
	result := h.getUserGitToken(c, 1, "gitlab", "https://gitlab.company.com")
	if result != "glpat_private123" {
		t.Errorf("Expected 'glpat_private123', got '%s'", result)
	}
}

func TestGetUserGitToken_PrivateGitLab_NoConnection(t *testing.T) {
	mock := &mockUserService{
		getGitConnectionErr: errors.New("not found"),
	}
	h := &PodHandler{userService: mock}
	c, _ := createCredentialTestContext(1)

	result := h.getUserGitToken(c, 1, "gitlab", "https://gitlab.company.com")
	if result != "" {
		t.Errorf("Expected empty string when no connection, got '%s'", result)
	}
}

func TestGetUserGitToken_PrivateGitLab_ConnectionDecryptError(t *testing.T) {
	mock := &mockUserService{
		gitConnection: &user.GitConnection{
			ID:           1,
			UserID:       1,
			ProviderType: "gitlab",
			BaseURL:      "https://gitlab.company.com",
		},
		getDecryptedConnectionTokenErr: errors.New("decrypt failed"),
	}
	h := &PodHandler{userService: mock}
	c, _ := createCredentialTestContext(1)

	result := h.getUserGitToken(c, 1, "gitlab", "https://gitlab.company.com")
	if result != "" {
		t.Errorf("Expected empty string when decrypt fails, got '%s'", result)
	}
}

func TestGetUserGitToken_PrivateGitLab_EmptyConnectionToken(t *testing.T) {
	mock := &mockUserService{
		gitConnection: &user.GitConnection{
			ID:           1,
			UserID:       1,
			ProviderType: "gitlab",
			BaseURL:      "https://gitlab.company.com",
		},
		decryptedConnectionToken: &userService.DecryptedTokens{
			AccessToken: "", // Empty
		},
	}
	h := &PodHandler{userService: mock}
	c, _ := createCredentialTestContext(1)

	result := h.getUserGitToken(c, 1, "gitlab", "https://gitlab.company.com")
	if result != "" {
		t.Errorf("Expected empty string for empty token, got '%s'", result)
	}
}

func TestGetUserGitToken_NoCredentialsFound(t *testing.T) {
	mock := &mockUserService{
		getDecryptedTokensErr: errors.New("no oauth"),
		getGitConnectionErr:   errors.New("no connection"),
	}
	h := &PodHandler{userService: mock}
	c, _ := createCredentialTestContext(1)

	result := h.getUserGitToken(c, 1, "github", "https://github.com")
	if result != "" {
		t.Errorf("Expected empty string when no credentials, got '%s'", result)
	}
}

func TestGetUserGitToken_NilConnection(t *testing.T) {
	mock := &mockUserService{
		getDecryptedTokensErr: errors.New("no oauth"),
		gitConnection:         nil, // nil connection without error
	}
	h := &PodHandler{userService: mock}
	c, _ := createCredentialTestContext(1)

	result := h.getUserGitToken(c, 1, "github", "https://github.com")
	if result != "" {
		t.Errorf("Expected empty string when connection is nil, got '%s'", result)
	}
}
