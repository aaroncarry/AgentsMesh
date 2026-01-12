package user

import (
	"time"
)

// GitCredential represents a user's Git credential for Git operations (clone/push/pull)
// Multiple credential types are supported:
// - runner_local: Use Runner machine's local git config (virtual, no credential stored)
// - oauth: Shared from Repository Provider (references provider)
// - pat: Personal Access Token
// - ssh_key: SSH private key
type GitCredential struct {
	ID     int64 `gorm:"primaryKey" json:"id"`
	UserID int64 `gorm:"not null;index" json:"user_id"`

	Name           string `gorm:"size:100;not null" json:"name"`
	CredentialType string `gorm:"size:20;not null" json:"credential_type"` // runner_local, oauth, pat, ssh_key

	// OAuth type: reference to Repository Provider
	RepositoryProviderID *int64 `gorm:"index" json:"repository_provider_id,omitempty"`

	// PAT type
	PATEncrypted *string `gorm:"type:text;column:pat_encrypted" json:"-"`

	// SSH Key type
	PublicKey           *string `gorm:"type:text" json:"public_key,omitempty"`
	PrivateKeyEncrypted *string `gorm:"type:text" json:"-"`
	Fingerprint         *string `gorm:"size:255" json:"fingerprint,omitempty"`

	// Host pattern for matching repositories (optional)
	HostPattern *string `gorm:"size:255" json:"host_pattern,omitempty"` // e.g., github.com, *, etc.

	// Status
	IsDefault bool `gorm:"not null;default:false" json:"is_default"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`

	// Associations
	User               *User               `gorm:"foreignKey:UserID" json:"user,omitempty"`
	RepositoryProvider *RepositoryProvider `gorm:"foreignKey:RepositoryProviderID" json:"repository_provider,omitempty"`
}

func (GitCredential) TableName() string {
	return "user_git_credentials"
}

// Credential types
const (
	CredentialTypeRunnerLocal = "runner_local"
	CredentialTypeOAuth       = "oauth"
	CredentialTypePAT         = "pat"
	CredentialTypeSSHKey      = "ssh_key"
)

// ValidCredentialTypes returns valid credential types
func ValidCredentialTypes() []string {
	return []string{CredentialTypeRunnerLocal, CredentialTypeOAuth, CredentialTypePAT, CredentialTypeSSHKey}
}

// IsValidCredentialType checks if the credential type is valid
func IsValidCredentialType(credentialType string) bool {
	for _, t := range ValidCredentialTypes() {
		if t == credentialType {
			return true
		}
	}
	return false
}

// GitCredentialResponse is the API response for a Git credential
type GitCredentialResponse struct {
	ID                   int64   `json:"id"`
	Name                 string  `json:"name"`
	CredentialType       string  `json:"credential_type"`
	RepositoryProviderID *int64  `json:"repository_provider_id,omitempty"`
	ProviderName         *string `json:"provider_name,omitempty"` // Populated from RepositoryProvider
	PublicKey            *string `json:"public_key,omitempty"`
	Fingerprint          *string `json:"fingerprint,omitempty"`
	HostPattern          *string `json:"host_pattern,omitempty"`
	IsDefault            bool    `json:"is_default"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
}

// ToResponse converts GitCredential to API response
func (c *GitCredential) ToResponse() *GitCredentialResponse {
	resp := &GitCredentialResponse{
		ID:                   c.ID,
		Name:                 c.Name,
		CredentialType:       c.CredentialType,
		RepositoryProviderID: c.RepositoryProviderID,
		PublicKey:            c.PublicKey,
		Fingerprint:          c.Fingerprint,
		HostPattern:          c.HostPattern,
		IsDefault:            c.IsDefault,
		CreatedAt:            c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            c.UpdatedAt.Format(time.RFC3339),
	}

	// Populate provider name if available
	if c.RepositoryProvider != nil {
		resp.ProviderName = &c.RepositoryProvider.Name
	}

	return resp
}

// RunnerLocalCredentialResponse returns a virtual "Runner Local" credential response
// This is not stored in the database, but represents the default option
func RunnerLocalCredentialResponse() *GitCredentialResponse {
	return &GitCredentialResponse{
		ID:             0,
		Name:           "Runner Local",
		CredentialType: CredentialTypeRunnerLocal,
		IsDefault:      false, // Will be set based on user preference
		CreatedAt:      "",
		UpdatedAt:      "",
	}
}
