package gitprovider

import (
	"time"
)

// Provider types
const (
	ProviderTypeGitHub = "github"
	ProviderTypeGitLab = "gitlab"
	ProviderTypeGitee  = "gitee"
	ProviderTypeSSH    = "ssh" // SSH-based Git server (no API)
)

// GitProvider represents a configured Git provider for an organization
type GitProvider struct {
	ID             int64  `gorm:"primaryKey" json:"id"`
	OrganizationID int64  `gorm:"not null;index" json:"organization_id"`
	ProviderType   string `gorm:"size:50;not null" json:"provider_type"` // gitlab, github, gitee, ssh
	Name           string `gorm:"size:100;not null" json:"name"`
	BaseURL        string `gorm:"size:255;not null" json:"base_url"`

	ClientID              *string `gorm:"size:255" json:"client_id,omitempty"`
	ClientSecretEncrypted *string `gorm:"type:text" json:"-"`
	BotTokenEncrypted     *string `gorm:"type:text" json:"-"`

	// SSHKeyID is used for SSH type providers to reference the SSH key
	SSHKeyID *int64 `gorm:"index" json:"ssh_key_id,omitempty"`

	IsDefault bool `gorm:"not null;default:false" json:"is_default"`
	IsActive  bool `gorm:"not null;default:true" json:"is_active"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`

	// Note: Repositories no longer have GitProviderID foreign key
	// Repositories are now self-contained with provider_type and provider_base_url
}

// IsSSHProvider returns true if this is an SSH-based provider
func (g *GitProvider) IsSSHProvider() bool {
	return g.ProviderType == ProviderTypeSSH
}

// HasAPIAccess returns true if this provider supports API access
func (g *GitProvider) HasAPIAccess() bool {
	return g.ProviderType != ProviderTypeSSH
}

func (GitProvider) TableName() string {
	return "git_providers"
}

// Repository represents a Git repository configured in the system
// Self-contained design: repository stores all necessary info, no git_provider_id dependency
type Repository struct {
	ID             int64 `gorm:"primaryKey" json:"id"`
	OrganizationID int64 `gorm:"not null;index" json:"organization_id"`

	// Provider info (self-contained, no foreign key to git_providers)
	ProviderType    string `gorm:"size:50;not null" json:"provider_type"`      // github, gitlab, gitee, generic
	ProviderBaseURL string `gorm:"size:255;not null" json:"provider_base_url"` // https://github.com, https://gitlab.company.com
	CloneURL        string `gorm:"size:500" json:"clone_url"`                  // Full clone URL

	ExternalID    string  `gorm:"size:255;not null" json:"external_id"`
	Name          string  `gorm:"size:255;not null" json:"name"`
	FullPath      string  `gorm:"size:500;not null" json:"full_path"`
	DefaultBranch string  `gorm:"size:100;default:'main'" json:"default_branch"`
	TicketPrefix  *string `gorm:"size:10" json:"ticket_prefix,omitempty"`

	// Visibility: "organization" (all members can see), "private" (only importer can see)
	Visibility       string `gorm:"size:20;not null;default:'organization'" json:"visibility"`
	ImportedByUserID *int64 `gorm:"index" json:"imported_by_user_id,omitempty"` // User who imported this repo

	IsActive bool `gorm:"not null;default:true" json:"is_active"`

	CreatedAt time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time  `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"` // Soft delete support
}

func (Repository) TableName() string {
	return "repositories"
}
