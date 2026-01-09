package user

import (
	"context"
	"errors"
	"time"

	"github.com/anthropics/agentmesh/backend/internal/domain/user"
	"github.com/anthropics/agentmesh/backend/pkg/crypto"
	"gorm.io/gorm"
)

var (
	ErrConnectionNotFound      = errors.New("git connection not found")
	ErrConnectionAlreadyExists = errors.New("git connection already exists for this provider and URL")
	ErrInvalidConnectionID     = errors.New("invalid connection ID format")
)

// CreateGitConnectionRequest represents a request to create a Git connection
type CreateGitConnectionRequest struct {
	ProviderType    string
	ProviderName    string
	BaseURL         string
	AuthType        string // "pat" or "ssh"
	AccessToken     string // Plain text, will be encrypted
	SSHPrivateKey   string // Plain text, will be encrypted
	ExternalUserID  string
	ExternalUsername string
	ExternalAvatarURL string
}

// CreateGitConnection creates a new Git connection for a user
func (s *Service) CreateGitConnection(ctx context.Context, userID int64, req *CreateGitConnectionRequest) (*user.GitConnection, error) {
	// Check if connection already exists
	var existing user.GitConnection
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND provider_type = ? AND base_url = ?", userID, req.ProviderType, req.BaseURL).
		First(&existing).Error
	if err == nil {
		return nil, ErrConnectionAlreadyExists
	}

	conn := &user.GitConnection{
		UserID:       userID,
		ProviderType: req.ProviderType,
		ProviderName: req.ProviderName,
		BaseURL:      req.BaseURL,
		AuthType:     req.AuthType,
		IsActive:     true,
	}

	// Set external user info
	if req.ExternalUserID != "" {
		conn.ExternalUserID = &req.ExternalUserID
	}
	if req.ExternalUsername != "" {
		conn.ExternalUsername = &req.ExternalUsername
	}
	if req.ExternalAvatarURL != "" {
		conn.ExternalAvatarURL = &req.ExternalAvatarURL
	}

	// Encrypt credentials
	if s.encryptionKey != "" {
		if req.AccessToken != "" {
			encrypted, err := crypto.EncryptWithKey(req.AccessToken, s.encryptionKey)
			if err != nil {
				return nil, err
			}
			conn.AccessTokenEncrypted = &encrypted
		}
		if req.SSHPrivateKey != "" {
			encrypted, err := crypto.EncryptWithKey(req.SSHPrivateKey, s.encryptionKey)
			if err != nil {
				return nil, err
			}
			conn.SSHPrivateKeyEncrypted = &encrypted
		}
	} else {
		// No encryption key - store as-is (not recommended)
		if req.AccessToken != "" {
			conn.AccessTokenEncrypted = &req.AccessToken
		}
		if req.SSHPrivateKey != "" {
			conn.SSHPrivateKeyEncrypted = &req.SSHPrivateKey
		}
	}

	if err := s.db.WithContext(ctx).Create(conn).Error; err != nil {
		return nil, err
	}

	return conn, nil
}

// GetGitConnection returns a Git connection by ID
func (s *Service) GetGitConnection(ctx context.Context, userID, connectionID int64) (*user.GitConnection, error) {
	var conn user.GitConnection
	err := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", connectionID, userID).
		First(&conn).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConnectionNotFound
		}
		return nil, err
	}
	return &conn, nil
}

// GetGitConnectionByProviderAndURL returns a Git connection by provider type and base URL
func (s *Service) GetGitConnectionByProviderAndURL(ctx context.Context, userID int64, providerType, baseURL string) (*user.GitConnection, error) {
	var conn user.GitConnection
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND provider_type = ? AND base_url = ?", userID, providerType, baseURL).
		First(&conn).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConnectionNotFound
		}
		return nil, err
	}
	return &conn, nil
}

// ListGitConnections returns all Git connections for a user
func (s *Service) ListGitConnections(ctx context.Context, userID int64) ([]*user.GitConnection, error) {
	var connections []*user.GitConnection
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&connections).Error
	return connections, err
}

// UpdateGitConnection updates a Git connection
func (s *Service) UpdateGitConnection(ctx context.Context, userID, connectionID int64, updates map[string]interface{}) (*user.GitConnection, error) {
	// Verify ownership
	conn, err := s.GetGitConnection(ctx, userID, connectionID)
	if err != nil {
		return nil, err
	}

	// Handle token encryption if updating tokens
	if token, ok := updates["access_token"].(string); ok && token != "" {
		if s.encryptionKey != "" {
			encrypted, err := crypto.EncryptWithKey(token, s.encryptionKey)
			if err != nil {
				return nil, err
			}
			updates["access_token_encrypted"] = encrypted
		} else {
			updates["access_token_encrypted"] = token
		}
		delete(updates, "access_token")
	}

	if sshKey, ok := updates["ssh_private_key"].(string); ok && sshKey != "" {
		if s.encryptionKey != "" {
			encrypted, err := crypto.EncryptWithKey(sshKey, s.encryptionKey)
			if err != nil {
				return nil, err
			}
			updates["ssh_private_key_encrypted"] = encrypted
		} else {
			updates["ssh_private_key_encrypted"] = sshKey
		}
		delete(updates, "ssh_private_key")
	}

	if err := s.db.WithContext(ctx).Model(conn).Updates(updates).Error; err != nil {
		return nil, err
	}

	return s.GetGitConnection(ctx, userID, connectionID)
}

// DeleteGitConnection deletes a Git connection
func (s *Service) DeleteGitConnection(ctx context.Context, userID, connectionID int64) error {
	result := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", connectionID, userID).
		Delete(&user.GitConnection{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrConnectionNotFound
	}
	return nil
}

// GetDecryptedConnectionToken retrieves and decrypts tokens for a Git connection
func (s *Service) GetDecryptedConnectionToken(ctx context.Context, userID, connectionID int64) (*DecryptedTokens, error) {
	conn, err := s.GetGitConnection(ctx, userID, connectionID)
	if err != nil {
		return nil, err
	}

	tokens := &DecryptedTokens{}

	if s.encryptionKey != "" {
		if conn.AccessTokenEncrypted != nil && *conn.AccessTokenEncrypted != "" {
			decrypted, err := crypto.DecryptWithKey(*conn.AccessTokenEncrypted, s.encryptionKey)
			if err != nil {
				return nil, err
			}
			tokens.AccessToken = decrypted
		}
	} else {
		if conn.AccessTokenEncrypted != nil {
			tokens.AccessToken = *conn.AccessTokenEncrypted
		}
	}

	return tokens, nil
}

// UpdateConnectionLastUsed updates the last_used_at timestamp
func (s *Service) UpdateConnectionLastUsed(ctx context.Context, connectionID int64) error {
	now := time.Now()
	return s.db.WithContext(ctx).
		Model(&user.GitConnection{}).
		Where("id = ?", connectionID).
		Update("last_used_at", now).Error
}

// GetAllUserGitConnections returns all Git connections including OAuth identities
// This merges user_identities (OAuth) and user_git_connections (PAT/SSH)
func (s *Service) GetAllUserGitConnections(ctx context.Context, userID int64) ([]*user.GitConnectionResponse, error) {
	var result []*user.GitConnectionResponse

	// 1. Get OAuth identities (from user_identities)
	identities, err := s.ListIdentities(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, identity := range identities {
		// Only include Git providers (github, gitlab, gitee), skip google
		if identity.Provider == "github" || identity.Provider == "gitlab" || identity.Provider == "gitee" {
			result = append(result, user.IdentityToConnectionResponse(identity))
		}
	}

	// 2. Get manual connections (from user_git_connections)
	connections, err := s.ListGitConnections(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, conn := range connections {
		result = append(result, conn.ToResponse())
	}

	return result, nil
}
