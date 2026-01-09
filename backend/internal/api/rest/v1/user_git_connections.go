package v1

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/anthropics/agentmesh/backend/internal/infra/git"
	"github.com/anthropics/agentmesh/backend/internal/middleware"
	"github.com/anthropics/agentmesh/backend/internal/service/user"
	"github.com/gin-gonic/gin"
)

// UserGitConnectionHandler handles user Git connection requests
type UserGitConnectionHandler struct {
	userService *user.Service
}

// NewUserGitConnectionHandler creates a new user Git connection handler
func NewUserGitConnectionHandler(userSvc *user.Service) *UserGitConnectionHandler {
	return &UserGitConnectionHandler{
		userService: userSvc,
	}
}

// RegisterRoutes registers user Git connection routes
func (h *UserGitConnectionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	connections := rg.Group("/user/git-connections")
	{
		connections.GET("", h.ListConnections)
		connections.POST("", h.CreateConnection)
		connections.GET("/:id", h.GetConnection)
		connections.PUT("/:id", h.UpdateConnection)
		connections.DELETE("/:id", h.DeleteConnection)
		connections.GET("/:id/repositories", h.ListRepositories)
	}
}

// ListConnections lists all Git connections for the current user
// GET /api/v1/user/git-connections
func (h *UserGitConnectionHandler) ListConnections(c *gin.Context) {
	userID := middleware.GetUserID(c)

	connections, err := h.userService.GetAllUserGitConnections(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list connections"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"connections": connections})
}

// CreateConnectionRequest represents a request to create a Git connection
type CreateConnectionRequest struct {
	ProviderType string `json:"provider_type" binding:"required"`
	ProviderName string `json:"provider_name" binding:"required"`
	BaseURL      string `json:"base_url" binding:"required"`
	AuthType     string `json:"auth_type"` // "pat" or "ssh", defaults to "pat"
	AccessToken  string `json:"access_token"`
	SSHPrivateKey string `json:"ssh_private_key"`
}

// CreateConnection creates a new Git connection
// POST /api/v1/user/git-connections
func (h *UserGitConnectionHandler) CreateConnection(c *gin.Context) {
	var req CreateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default auth type
	if req.AuthType == "" {
		req.AuthType = "pat"
	}

	// Validate auth type and credentials
	if req.AuthType == "pat" && req.AccessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "access_token is required for PAT authentication"})
		return
	}
	if req.AuthType == "ssh" && req.SSHPrivateKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ssh_private_key is required for SSH authentication"})
		return
	}

	userID := middleware.GetUserID(c)

	// TODO: Validate the token by calling the Git provider API
	// This should fetch user info and verify the token is valid

	conn, err := h.userService.CreateGitConnection(c.Request.Context(), userID, &user.CreateGitConnectionRequest{
		ProviderType:  req.ProviderType,
		ProviderName:  req.ProviderName,
		BaseURL:       req.BaseURL,
		AuthType:      req.AuthType,
		AccessToken:   req.AccessToken,
		SSHPrivateKey: req.SSHPrivateKey,
	})
	if err != nil {
		if err == user.ErrConnectionAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "Connection already exists for this provider and URL"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create connection"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"connection": conn.ToResponse()})
}

// GetConnection returns a single Git connection
// GET /api/v1/user/git-connections/:id
func (h *UserGitConnectionHandler) GetConnection(c *gin.Context) {
	userID := middleware.GetUserID(c)

	connectionID, err := parseConnectionID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	conn, err := h.userService.GetGitConnection(c.Request.Context(), userID, connectionID)
	if err != nil {
		if err == user.ErrConnectionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get connection"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"connection": conn.ToResponse()})
}

// UpdateConnectionRequest represents a request to update a Git connection
type UpdateConnectionRequest struct {
	ProviderName string `json:"provider_name"`
	AccessToken  string `json:"access_token"`
	SSHPrivateKey string `json:"ssh_private_key"`
	IsActive     *bool  `json:"is_active"`
}

// UpdateConnection updates a Git connection
// PUT /api/v1/user/git-connections/:id
func (h *UserGitConnectionHandler) UpdateConnection(c *gin.Context) {
	userID := middleware.GetUserID(c)

	connectionID, err := parseConnectionID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	var req UpdateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.ProviderName != "" {
		updates["provider_name"] = req.ProviderName
	}
	if req.AccessToken != "" {
		updates["access_token"] = req.AccessToken
	}
	if req.SSHPrivateKey != "" {
		updates["ssh_private_key"] = req.SSHPrivateKey
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No updates provided"})
		return
	}

	conn, err := h.userService.UpdateGitConnection(c.Request.Context(), userID, connectionID, updates)
	if err != nil {
		if err == user.ErrConnectionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update connection"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"connection": conn.ToResponse()})
}

// DeleteConnection deletes a Git connection
// DELETE /api/v1/user/git-connections/:id
func (h *UserGitConnectionHandler) DeleteConnection(c *gin.Context) {
	userID := middleware.GetUserID(c)

	connectionID, err := parseConnectionID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	err = h.userService.DeleteGitConnection(c.Request.Context(), userID, connectionID)
	if err != nil {
		if err == user.ErrConnectionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete connection"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Connection deleted"})
}

// parseConnectionID parses a connection ID from string
// Supports formats: "123" or "connection:123"
func parseConnectionID(idStr string) (int64, error) {
	// Remove "connection:" prefix if present
	if strings.HasPrefix(idStr, "connection:") {
		idStr = strings.TrimPrefix(idStr, "connection:")
	}
	return strconv.ParseInt(idStr, 10, 64)
}

// RepositoryResponse represents a repository in the API response
type RepositoryResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	FullPath      string `json:"full_path"`
	Description   string `json:"description"`
	DefaultBranch string `json:"default_branch"`
	Visibility    string `json:"visibility"`
	CloneURL      string `json:"clone_url"`
	SSHCloneURL   string `json:"ssh_clone_url"`
	WebURL        string `json:"web_url"`
}

// ListRepositories lists repositories accessible through a Git connection
// GET /api/v1/user/git-connections/:id/repositories
func (h *UserGitConnectionHandler) ListRepositories(c *gin.Context) {
	userID := middleware.GetUserID(c)
	connectionID := c.Param("id")

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	// Get provider type, base URL, and access token based on connection ID format
	var providerType, baseURL, accessToken string
	var err error

	if strings.HasPrefix(connectionID, "oauth:") {
		// OAuth identity: format is "oauth:github" or "oauth:gitlab"
		provider := strings.TrimPrefix(connectionID, "oauth:")
		providerType = provider

		// Get base URL for OAuth providers
		switch provider {
		case "github":
			baseURL = "https://api.github.com"
		case "gitlab":
			baseURL = "https://gitlab.com"
		case "gitee":
			baseURL = "https://gitee.com"
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported OAuth provider"})
			return
		}

		// Get decrypted OAuth token
		tokens, err := h.userService.GetDecryptedTokens(c.Request.Context(), userID, provider)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "OAuth identity not found or token unavailable"})
			return
		}
		accessToken = tokens.AccessToken
	} else {
		// Personal connection: format is "connection:123" or just "123"
		connID, parseErr := parseConnectionID(connectionID)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
			return
		}

		// Get connection details
		conn, connErr := h.userService.GetGitConnection(c.Request.Context(), userID, connID)
		if connErr != nil {
			if connErr == user.ErrConnectionNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get connection"})
			return
		}

		providerType = conn.ProviderType
		baseURL = conn.BaseURL

		// Get decrypted PAT token
		decryptedTokens, tokenErr := h.userService.GetDecryptedConnectionToken(c.Request.Context(), userID, connID)
		if tokenErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt access token"})
			return
		}
		accessToken = decryptedTokens.AccessToken
	}

	if accessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No access token available for this connection"})
		return
	}

	// Create git provider
	provider, err := git.NewProvider(providerType, baseURL, accessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create git provider: " + err.Error()})
		return
	}

	// Fetch repositories
	var projects []*git.Project
	if search != "" {
		projects, err = provider.SearchProjects(c.Request.Context(), search, page, perPage)
	} else {
		projects, err = provider.ListProjects(c.Request.Context(), page, perPage)
	}

	if err != nil {
		if err == git.ErrUnauthorized {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Access token is invalid or expired"})
			return
		}
		if err == git.ErrRateLimited {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limited by git provider"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch repositories: " + err.Error()})
		return
	}

	// Convert to response format
	repositories := make([]*RepositoryResponse, len(projects))
	for i, p := range projects {
		repositories[i] = &RepositoryResponse{
			ID:            p.ID,
			Name:          p.Name,
			FullPath:      p.FullPath,
			Description:   p.Description,
			DefaultBranch: p.DefaultBranch,
			Visibility:    p.Visibility,
			CloneURL:      p.CloneURL,
			SSHCloneURL:   p.SSHCloneURL,
			WebURL:        p.WebURL,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"repositories": repositories,
		"page":         page,
		"per_page":     perPage,
	})
}
