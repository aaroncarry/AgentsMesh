package v1

import (
	"log"
	"net/http"
	"strconv"

	"github.com/anthropics/agentmesh/backend/internal/middleware"
	"github.com/anthropics/agentmesh/backend/internal/service/agent"
	"github.com/anthropics/agentmesh/backend/internal/service/agentpod"
	"github.com/anthropics/agentmesh/backend/internal/service/gitprovider"
	"github.com/anthropics/agentmesh/backend/internal/service/repository"
	"github.com/anthropics/agentmesh/backend/internal/service/runner"
	"github.com/anthropics/agentmesh/backend/internal/service/sshkey"
	"github.com/anthropics/agentmesh/backend/internal/service/ticket"
	"github.com/anthropics/agentmesh/backend/internal/service/user"
	"github.com/gin-gonic/gin"
)

// PodHandler handles pod-related requests
type PodHandler struct {
	podService         *agentpod.PodService
	runnerService      *runner.Service
	agentService       *agent.Service
	repositoryService  *repository.Service
	ticketService      *ticket.Service
	gitProviderService *gitprovider.Service
	sshKeyService      *sshkey.Service
	userService        *user.Service // For user credential retrieval (权限跟人走)
	runnerConnMgr      *runner.ConnectionManager
	podCoordinator     *runner.PodCoordinator
	terminalRouter     interface{} // *runner.TerminalRouter, optional
}

// PodHandlerOption is a functional option for configuring PodHandler
type PodHandlerOption func(*PodHandler)

// WithRunnerConnectionManager sets the runner connection manager
func WithRunnerConnectionManager(cm *runner.ConnectionManager) PodHandlerOption {
	return func(h *PodHandler) {
		h.runnerConnMgr = cm
	}
}

// WithPodCoordinator sets the pod coordinator
func WithPodCoordinator(pc *runner.PodCoordinator) PodHandlerOption {
	return func(h *PodHandler) {
		h.podCoordinator = pc
	}
}

// WithTerminalRouter sets the terminal router
func WithTerminalRouter(tr interface{}) PodHandlerOption {
	return func(h *PodHandler) {
		h.terminalRouter = tr
	}
}

// WithRepositoryService sets the repository service
func WithRepositoryService(rs *repository.Service) PodHandlerOption {
	return func(h *PodHandler) {
		h.repositoryService = rs
	}
}

// WithTicketService sets the ticket service
func WithTicketService(ts *ticket.Service) PodHandlerOption {
	return func(h *PodHandler) {
		h.ticketService = ts
	}
}

// WithGitProviderService sets the git provider service
func WithGitProviderService(gps *gitprovider.Service) PodHandlerOption {
	return func(h *PodHandler) {
		h.gitProviderService = gps
	}
}

// WithSSHKeyService sets the SSH key service
func WithSSHKeyService(sks *sshkey.Service) PodHandlerOption {
	return func(h *PodHandler) {
		h.sshKeyService = sks
	}
}

// WithUserService sets the user service for credential retrieval (权限跟人走)
func WithUserService(us *user.Service) PodHandlerOption {
	return func(h *PodHandler) {
		h.userService = us
	}
}

// NewPodHandler creates a new pod handler with required dependencies and optional configurations
func NewPodHandler(podService *agentpod.PodService, runnerService *runner.Service, agentService *agent.Service, opts ...PodHandlerOption) *PodHandler {
	h := &PodHandler{
		podService:    podService,
		runnerService: runnerService,
		agentService:  agentService,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}


// ListPodsRequest represents pod list request
type ListPodsRequest struct {
	Status string `form:"status"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

// ListPods lists pods
// GET /api/v1/organizations/:slug/pods
func (h *PodHandler) ListPods(c *gin.Context) {
	var req ListPodsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenant := middleware.GetTenant(c)

	limit := req.Limit
	if limit == 0 {
		limit = 20
	}

	pods, total, err := h.podService.ListPods(
		c.Request.Context(),
		tenant.OrganizationID,
		nil, // TeamID is deprecated
		req.Status,
		limit,
		req.Offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list pods"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pods":   pods,
		"total":  total,
		"limit":  limit,
		"offset": req.Offset,
	})
}

// CreatePodRequest represents pod creation request
type CreatePodRequest struct {
	RunnerID          int64   `json:"runner_id" binding:"required"`
	AgentTypeID       *int64  `json:"agent_type_id"`
	CustomAgentTypeID *int64  `json:"custom_agent_type_id"`
	RepositoryID      *int64  `json:"repository_id"`
	RepositoryURL     *string `json:"repository_url"`      // Direct repository URL (takes precedence over repository_id)
	TicketID          *int64  `json:"ticket_id"`
	TicketIdentifier  *string `json:"ticket_identifier"`   // Direct ticket identifier (takes precedence over ticket_id)
	InitialPrompt     string  `json:"initial_prompt"`
	BranchName        *string `json:"branch_name"`
	PermissionMode    *string `json:"permission_mode"`     // "plan", "default", or "bypassPermissions"

	// PluginConfig allows advanced users to pass additional configuration to sandbox plugins
	// Fields: init_script, init_timeout, env_vars
	PluginConfig map[string]interface{} `json:"plugin_config"`
}

// CreatePod creates a new pod
// POST /api/v1/organizations/:slug/pods
func (h *PodHandler) CreatePod(c *gin.Context) {
	var req CreatePodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenant := middleware.GetTenant(c)

	// Create pod record in database
	pod, err := h.podService.CreatePod(c.Request.Context(), &agentpod.CreatePodRequest{
		OrganizationID:    tenant.OrganizationID,
		RunnerID:          req.RunnerID,
		AgentTypeID:       req.AgentTypeID,
		CustomAgentTypeID: req.CustomAgentTypeID,
		RepositoryID:      req.RepositoryID,
		TicketID:          req.TicketID,
		CreatedByID:       tenant.UserID,
		InitialPrompt:     req.InitialPrompt,
		BranchName:        req.BranchName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pod"})
		return
	}

	// Send create_pod command to runner via PodCoordinator
	if h.podCoordinator != nil {
		// Get permission mode from request or pod settings (default to "plan")
		permissionMode := "plan"
		if req.PermissionMode != nil {
			permissionMode = *req.PermissionMode
		} else if pod.PermissionMode != nil {
			permissionMode = *pod.PermissionMode
		}

		// Build PluginConfig for Runner's Sandbox plugins
		pluginConfig := h.buildPluginConfig(c, &req)

		createReq := &runner.CreatePodRequest{
			PodKey:         pod.PodKey,
			InitialCommand: "claude", // Default command to run Claude Code CLI
			InitialPrompt:  req.InitialPrompt,
			PermissionMode: permissionMode,
			PluginConfig:   pluginConfig,
		}

		// Log the request
		log.Printf("[pods] Sending create_pod to runner %d for pod %s with plugin_config: %v",
			req.RunnerID, pod.PodKey, pluginConfig)

		if err := h.podCoordinator.CreatePod(c.Request.Context(), req.RunnerID, createReq); err != nil {
			// Log the error but don't fail - pod is created, runner might be offline
			log.Printf("[pods] Failed to send create_pod: %v", err)
			c.JSON(http.StatusCreated, gin.H{
				"pod":     pod,
				"warning": "Pod created but runner communication failed: " + err.Error(),
			})
			return
		}
		log.Printf("[pods] create_pod sent successfully for pod %s", pod.PodKey)
	} else {
		log.Printf("[pods] PodCoordinator is nil, cannot send create_pod command")
	}

	c.JSON(http.StatusCreated, gin.H{"pod": pod})
}

// GetPod returns pod by key
// GET /api/v1/organizations/:slug/pods/:key
func (h *PodHandler) GetPod(c *gin.Context) {
	podKey := c.Param("key")

	pod, err := h.podService.GetPod(c.Request.Context(), podKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pod not found"})
		return
	}

	tenant := middleware.GetTenant(c)
	if pod.OrganizationID != tenant.OrganizationID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// All organization members can access pods (Team-based access control removed)
	c.JSON(http.StatusOK, gin.H{"pod": pod})
}

// TerminatePod terminates a pod
// POST /api/v1/organizations/:slug/pods/:key/terminate
func (h *PodHandler) TerminatePod(c *gin.Context) {
	podKey := c.Param("key")

	pod, err := h.podService.GetPod(c.Request.Context(), podKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pod not found"})
		return
	}

	tenant := middleware.GetTenant(c)
	if pod.OrganizationID != tenant.OrganizationID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Only creator or admin can terminate
	if pod.CreatedByID != tenant.UserID && tenant.UserRole == "member" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only pod creator or admin can terminate"})
		return
	}

	if err := h.podService.TerminatePod(c.Request.Context(), podKey); err != nil {
		if err == agentpod.ErrPodTerminated {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Pod already terminated"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to terminate pod"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pod terminated"})
}

// GetPodConnection returns connection info for pod
// GET /api/v1/organizations/:slug/pods/:key/connect
func (h *PodHandler) GetPodConnection(c *gin.Context) {
	podKey := c.Param("key")

	pod, err := h.podService.GetPod(c.Request.Context(), podKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pod not found"})
		return
	}

	tenant := middleware.GetTenant(c)
	if pod.OrganizationID != tenant.OrganizationID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if !pod.IsActive() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pod is not active"})
		return
	}

	// Return WebSocket connection URL
	c.JSON(http.StatusOK, gin.H{
		"pod_key": podKey,
		"ws_url":  "/api/v1/ws/terminal/" + podKey,
		"status":  pod.Status,
	})
}

// ListPodsByTicket lists pods for a ticket
// GET /api/v1/organizations/:slug/tickets/:id/pods
func (h *PodHandler) ListPodsByTicket(c *gin.Context) {
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	pods, err := h.podService.GetPodsByTicket(c.Request.Context(), ticketID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list pods"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pods": pods})
}

// GetConnectionInfo returns connection info for pod (alias for GetPodConnection)
// GET /api/v1/organizations/:slug/pods/:key/connect
func (h *PodHandler) GetConnectionInfo(c *gin.Context) {
	h.GetPodConnection(c)
}

// SendPromptRequest represents prompt sending request
type SendPromptRequest struct {
	Prompt string `json:"prompt" binding:"required"`
}

// SendPrompt sends a prompt to the pod
// POST /api/v1/organizations/:slug/pods/:key/send-prompt
func (h *PodHandler) SendPrompt(c *gin.Context) {
	podKey := c.Param("key")

	var req SendPromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pod, err := h.podService.GetPod(c.Request.Context(), podKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pod not found"})
		return
	}

	tenant := middleware.GetTenant(c)
	if pod.OrganizationID != tenant.OrganizationID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if !pod.IsActive() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pod is not active"})
		return
	}

	// TODO: Implement WebSocket-based prompt sending to runner
	// For now, return not implemented
	_ = req.Prompt
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Prompt sending via REST not implemented. Use WebSocket terminal."})
}

// TerminalRouterInterface defines the interface for terminal router operations
type TerminalRouterInterface interface {
	GetRecentOutput(podKey string, lines int, raw bool) []byte
	GetScreenSnapshot(podKey string) string
	GetCursorPosition(podKey string) (row, col int)
	GetAllScrollbackData(podKey string) []byte
	RouteInput(podKey string, data []byte) error
	RouteResize(podKey string, cols, rows int) error
}

// TerminalOutputResponse matches Runner's tools.TerminalOutput structure
type TerminalOutputResponse struct {
	PodKey     string `json:"pod_key"`
	Output     string `json:"output"`
	Screen     string `json:"screen,omitempty"`
	CursorX    int    `json:"cursor_x"`
	CursorY    int    `json:"cursor_y"`
	TotalLines int    `json:"total_lines"`
	HasMore    bool   `json:"has_more"`
}

// ObserveTerminalRequest represents terminal observation request
type ObserveTerminalRequest struct {
	Lines         int  `form:"lines"`
	Raw           bool `form:"raw"`            // If true, return raw output; otherwise return processed output
	IncludeScreen bool `form:"include_screen"` // If true, include current screen snapshot
}

// ObserveTerminal returns recent terminal output for observation
// GET /api/v1/organizations/:slug/pods/:key/terminal/observe
func (h *PodHandler) ObserveTerminal(c *gin.Context) {
	podKey := c.Param("key")

	var req ObserveTerminalRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pod, err := h.podService.GetPod(c.Request.Context(), podKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pod not found"})
		return
	}

	tenant := middleware.GetTenant(c)
	if pod.OrganizationID != tenant.OrganizationID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// All organization members can access pods (Team-based access control removed)

	// Get terminal output from router if available
	if h.terminalRouter == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Terminal router not available"})
		return
	}

	tr, ok := h.terminalRouter.(TerminalRouterInterface)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Terminal router interface not implemented"})
		return
	}

	lines := req.Lines
	if lines <= 0 {
		lines = 100 // Default to last 100 lines
	}

	var output []byte
	if lines == -1 {
		// Get all raw scrollback data
		output = tr.GetAllScrollbackData(podKey)
	} else {
		// Get recent output (processed by default, raw if requested)
		output = tr.GetRecentOutput(podKey, lines, req.Raw)
	}

	// Get cursor position from virtual terminal
	cursorRow, cursorCol := tr.GetCursorPosition(podKey)

	// Calculate total lines (rough estimate from output)
	totalLines := 0
	for _, b := range output {
		if b == '\n' {
			totalLines++
		}
	}
	if len(output) > 0 && output[len(output)-1] != '\n' {
		totalLines++ // Count last line if not ending with newline
	}

	// Build response matching Runner's TerminalOutput structure
	response := TerminalOutputResponse{
		PodKey:     podKey,
		Output:     string(output),
		CursorX:    cursorCol,
		CursorY:    cursorRow,
		TotalLines: totalLines,
		HasMore:    lines != -1 && totalLines >= lines,
	}

	// Include screen snapshot if requested
	if req.IncludeScreen {
		response.Screen = tr.GetScreenSnapshot(podKey)
	}

	c.JSON(http.StatusOK, response)
}

// TerminalInputRequest represents terminal input request
type TerminalInputRequest struct {
	Input string `json:"input" binding:"required"`
}

// SendTerminalInput sends input to the terminal
// POST /api/v1/organizations/:slug/pods/:key/terminal/input
func (h *PodHandler) SendTerminalInput(c *gin.Context) {
	podKey := c.Param("key")

	var req TerminalInputRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pod, err := h.podService.GetPod(c.Request.Context(), podKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pod not found"})
		return
	}

	tenant := middleware.GetTenant(c)
	if pod.OrganizationID != tenant.OrganizationID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if !pod.IsActive() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pod is not active"})
		return
	}

	// All organization members can access pods (Team-based access control removed)

	if h.terminalRouter == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Terminal router not available"})
		return
	}

	tr, ok := h.terminalRouter.(TerminalRouterInterface)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Terminal router interface not implemented"})
		return
	}

	if err := tr.RouteInput(podKey, []byte(req.Input)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send input: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Input sent"})
}

// TerminalResizeRequest represents terminal resize request
type TerminalResizeRequest struct {
	Cols int `json:"cols" binding:"required,min=1"`
	Rows int `json:"rows" binding:"required,min=1"`
}

// ResizeTerminal resizes the terminal
// POST /api/v1/organizations/:slug/pods/:key/terminal/resize
func (h *PodHandler) ResizeTerminal(c *gin.Context) {
	podKey := c.Param("key")

	var req TerminalResizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pod, err := h.podService.GetPod(c.Request.Context(), podKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pod not found"})
		return
	}

	tenant := middleware.GetTenant(c)
	if pod.OrganizationID != tenant.OrganizationID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if !pod.IsActive() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pod is not active"})
		return
	}

	if h.terminalRouter == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Terminal router not available"})
		return
	}

	tr, ok := h.terminalRouter.(TerminalRouterInterface)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Terminal router interface not implemented"})
		return
	}

	if err := tr.RouteResize(podKey, req.Cols, req.Rows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resize terminal: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Terminal resized"})
}

// buildPluginConfig builds the PluginConfig map for Runner's Sandbox plugins
// It resolves repository_id -> repository_url, ticket_id -> ticket_identifier
//
// Credential Strategy (权限跟人走):
// - Repository is self-contained with provider_type and provider_base_url
// - Credentials are obtained from the current user's OAuth identity or Git connections
// - If no user credentials available, Runner uses its local Git configuration
func (h *PodHandler) buildPluginConfig(c *gin.Context, req *CreatePodRequest) map[string]interface{} {
	config := make(map[string]interface{})
	userID := middleware.GetUserID(c)

	// 1. Resolve Repository URL
	// Priority: repository_url > repository_id
	if req.RepositoryURL != nil && *req.RepositoryURL != "" {
		config["repository_url"] = *req.RepositoryURL
	} else if req.RepositoryID != nil && h.repositoryService != nil {
		repo, err := h.repositoryService.GetByID(c.Request.Context(), *req.RepositoryID)
		if err == nil && repo != nil {
			// Use repository's self-contained clone URL
			config["repository_url"] = repo.CloneURL

			// Store provider info for potential use by Runner
			config["provider_type"] = repo.ProviderType
			config["provider_base_url"] = repo.ProviderBaseURL

			// Get credentials from current user (权限跟人走)
			// Try OAuth identity first, then PAT connections
			if h.userService != nil {
				token := h.getUserGitToken(c, userID, repo.ProviderType, repo.ProviderBaseURL)
				if token != "" {
					config["git_token"] = token
				}
				// If no token found, Runner will use its local Git configuration
			}
		}
	}

	// 2. Resolve Branch Name
	if req.BranchName != nil && *req.BranchName != "" {
		config["branch"] = *req.BranchName
	} else if req.RepositoryID != nil && h.repositoryService != nil {
		// Use repository's default branch if not specified
		repo, err := h.repositoryService.GetByID(c.Request.Context(), *req.RepositoryID)
		if err == nil && repo != nil && repo.DefaultBranch != "" {
			config["branch"] = repo.DefaultBranch
		}
	}

	// 3. Resolve Ticket Identifier
	// Priority: ticket_identifier > ticket_id
	if req.TicketIdentifier != nil && *req.TicketIdentifier != "" {
		config["ticket_identifier"] = *req.TicketIdentifier
	} else if req.TicketID != nil && h.ticketService != nil {
		t, err := h.ticketService.GetTicket(c.Request.Context(), *req.TicketID)
		if err == nil && t != nil {
			config["ticket_identifier"] = t.Identifier
		}
	}

	// 4. Merge user-provided PluginConfig (can override above values)
	if req.PluginConfig != nil {
		for k, v := range req.PluginConfig {
			config[k] = v
		}
	}

	return config
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
