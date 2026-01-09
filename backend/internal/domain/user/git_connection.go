package user

import (
	"strconv"
	"time"
)

// GitConnection represents a manually added Git provider connection for a user
// OAuth connections are stored in Identity, this is for PAT/SSH connections
type GitConnection struct {
	ID     int64 `gorm:"primaryKey" json:"id"`
	UserID int64 `gorm:"not null;index" json:"user_id"`

	// Provider info
	ProviderType string `gorm:"size:50;not null" json:"provider_type"` // github, gitlab, gitee, generic
	ProviderName string `gorm:"size:100;not null" json:"provider_name"` // User-defined name
	BaseURL      string `gorm:"size:255;not null" json:"base_url"`      // https://gitlab.company.com

	// Authentication
	AuthType               string  `gorm:"size:20;not null;default:'pat'" json:"auth_type"` // pat, ssh
	AccessTokenEncrypted   *string `gorm:"type:text" json:"-"`
	SSHPrivateKeyEncrypted *string `gorm:"type:text" json:"-"`

	// Provider user info
	ExternalUserID    *string `gorm:"size:255" json:"external_user_id,omitempty"`
	ExternalUsername  *string `gorm:"size:255" json:"external_username,omitempty"`
	ExternalAvatarURL *string `gorm:"type:text" json:"external_avatar_url,omitempty"`

	// Status
	IsActive   bool       `gorm:"not null;default:true" json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`

	// Associations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (GitConnection) TableName() string {
	return "user_git_connections"
}

// GitConnectionResponse is the API response for a Git connection
type GitConnectionResponse struct {
	ID           string `json:"id"`             // Format: "connection:123" or "oauth:github"
	Type         string `json:"type"`           // "oauth" or "personal"
	ProviderType string `json:"provider_type"`  // github, gitlab, gitee
	ProviderName string `json:"provider_name"`  // Display name
	BaseURL      string `json:"base_url"`       // https://github.com
	Username     string `json:"username"`       // Username on the platform
	AvatarURL    string `json:"avatar_url,omitempty"`
	AuthType     string `json:"auth_type,omitempty"` // pat, ssh (only for personal)
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
}

// ToResponse converts GitConnection to API response
func (c *GitConnection) ToResponse() *GitConnectionResponse {
	resp := &GitConnectionResponse{
		ID:           "connection:" + formatInt64(c.ID),
		Type:         "personal",
		ProviderType: c.ProviderType,
		ProviderName: c.ProviderName,
		BaseURL:      c.BaseURL,
		AuthType:     c.AuthType,
		IsActive:     c.IsActive,
		CreatedAt:    c.CreatedAt.Format(time.RFC3339),
	}
	if c.ExternalUsername != nil {
		resp.Username = *c.ExternalUsername
	}
	if c.ExternalAvatarURL != nil {
		resp.AvatarURL = *c.ExternalAvatarURL
	}
	return resp
}

// IdentityToConnectionResponse converts Identity to GitConnectionResponse
func IdentityToConnectionResponse(identity *Identity) *GitConnectionResponse {
	resp := &GitConnectionResponse{
		ID:           "oauth:" + identity.Provider,
		Type:         "oauth",
		ProviderType: identity.Provider,
		ProviderName: getProviderDisplayName(identity.Provider),
		BaseURL:      getProviderBaseURL(identity.Provider),
		IsActive:     true,
		CreatedAt:    identity.CreatedAt.Format(time.RFC3339),
	}
	if identity.ProviderUsername != nil {
		resp.Username = *identity.ProviderUsername
	}
	return resp
}

// Helper functions
func formatInt64(n int64) string {
	return strconv.FormatInt(n, 10)
}

func getProviderDisplayName(provider string) string {
	switch provider {
	case "github":
		return "GitHub"
	case "gitlab":
		return "GitLab"
	case "google":
		return "Google"
	case "gitee":
		return "Gitee"
	default:
		return provider
	}
}

func getProviderBaseURL(provider string) string {
	switch provider {
	case "github":
		return "https://github.com"
	case "gitlab":
		return "https://gitlab.com"
	case "gitee":
		return "https://gitee.com"
	default:
		return ""
	}
}
