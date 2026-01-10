package channel

import (
	"context"
	"errors"
	"time"

	"github.com/anthropics/agentmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentmesh/backend/internal/domain/channel"
	"gorm.io/gorm"
)

var (
	ErrChannelNotFound = errors.New("channel not found")
	ErrChannelArchived = errors.New("channel is archived")
	ErrDuplicateName   = errors.New("channel name already exists")
)

// Service handles channel operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new channel service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CreateChannelRequest represents a channel creation request
type CreateChannelRequest struct {
	OrganizationID   int64
	Name             string
	Description      *string
	RepositoryID     *int64
	TicketID         *int64
	CreatedByPod *string
	CreatedByUserID  *int64
}

// CreateChannel creates a new channel
func (s *Service) CreateChannel(ctx context.Context, req *CreateChannelRequest) (*channel.Channel, error) {
	// Check for duplicate name
	var existing channel.Channel
	if err := s.db.WithContext(ctx).
		Where("organization_id = ? AND name = ?", req.OrganizationID, req.Name).
		First(&existing).Error; err == nil {
		return nil, ErrDuplicateName
	}

	ch := &channel.Channel{
		OrganizationID:   req.OrganizationID,
		Name:             req.Name,
		Description:      req.Description,
		RepositoryID:     req.RepositoryID,
		TicketID:         req.TicketID,
		CreatedByPod: req.CreatedByPod,
		CreatedByUserID:  req.CreatedByUserID,
		IsArchived:       false,
	}

	if err := s.db.WithContext(ctx).Create(ch).Error; err != nil {
		return nil, err
	}

	return ch, nil
}

// GetChannel returns a channel by ID
func (s *Service) GetChannel(ctx context.Context, channelID int64) (*channel.Channel, error) {
	var ch channel.Channel
	if err := s.db.WithContext(ctx).First(&ch, channelID).Error; err != nil {
		return nil, ErrChannelNotFound
	}
	return &ch, nil
}

// GetChannelByName returns a channel by name within an organization
func (s *Service) GetChannelByName(ctx context.Context, orgID int64, name string) (*channel.Channel, error) {
	var ch channel.Channel
	if err := s.db.WithContext(ctx).
		Where("organization_id = ? AND name = ?", orgID, name).
		First(&ch).Error; err != nil {
		return nil, ErrChannelNotFound
	}
	return &ch, nil
}

// ListChannels returns channels for an organization
func (s *Service) ListChannels(ctx context.Context, orgID int64, includeArchived bool, limit, offset int) ([]*channel.Channel, int64, error) {
	query := s.db.WithContext(ctx).Model(&channel.Channel{}).Where("organization_id = ?", orgID)

	if !includeArchived {
		query = query.Where("is_archived = ?", false)
	}

	var total int64
	query.Count(&total)

	var channels []*channel.Channel
	if err := query.
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&channels).Error; err != nil {
		return nil, 0, err
	}

	return channels, total, nil
}

// UpdateChannel updates a channel
func (s *Service) UpdateChannel(ctx context.Context, channelID int64, name, description, document *string) (*channel.Channel, error) {
	ch, err := s.GetChannel(ctx, channelID)
	if err != nil {
		return nil, err
	}

	if ch.IsArchived {
		return nil, ErrChannelArchived
	}

	updates := make(map[string]interface{})
	if name != nil {
		updates["name"] = *name
	}
	if description != nil {
		updates["description"] = *description
	}
	if document != nil {
		updates["document"] = *document
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(ch).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return s.GetChannel(ctx, channelID)
}

// ArchiveChannel archives a channel
func (s *Service) ArchiveChannel(ctx context.Context, channelID int64) error {
	return s.db.WithContext(ctx).Model(&channel.Channel{}).
		Where("id = ?", channelID).
		Update("is_archived", true).Error
}

// UnarchiveChannel unarchives a channel
func (s *Service) UnarchiveChannel(ctx context.Context, channelID int64) error {
	return s.db.WithContext(ctx).Model(&channel.Channel{}).
		Where("id = ?", channelID).
		Update("is_archived", false).Error
}

// SendMessage sends a message to a channel
func (s *Service) SendMessage(ctx context.Context, channelID int64, senderPod *string, senderUserID *int64, messageType, content string, metadata channel.MessageMetadata) (*channel.Message, error) {
	ch, err := s.GetChannel(ctx, channelID)
	if err != nil {
		return nil, err
	}

	if ch.IsArchived {
		return nil, ErrChannelArchived
	}

	msg := &channel.Message{
		ChannelID:     channelID,
		SenderPod: senderPod,
		SenderUserID:  senderUserID,
		MessageType:   messageType,
		Content:       content,
		Metadata:      metadata,
	}

	if err := s.db.WithContext(ctx).Create(msg).Error; err != nil {
		return nil, err
	}

	// Update channel updated_at
	s.db.WithContext(ctx).Model(ch).Update("updated_at", time.Now())

	return msg, nil
}

// GetMessages returns messages for a channel
func (s *Service) GetMessages(ctx context.Context, channelID int64, before *time.Time, limit int) ([]*channel.Message, error) {
	query := s.db.WithContext(ctx).Where("channel_id = ?", channelID)

	if before != nil {
		query = query.Where("created_at < ?", *before)
	}

	var messages []*channel.Message
	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetChannelsByTicket returns channels for a ticket
func (s *Service) GetChannelsByTicket(ctx context.Context, ticketID int64) ([]*channel.Channel, error) {
	var channels []*channel.Channel
	if err := s.db.WithContext(ctx).
		Where("ticket_id = ?", ticketID).
		Find(&channels).Error; err != nil {
		return nil, err
	}
	return channels, nil
}

// Pod Binding operations

// CreateBinding creates a pod binding request
func (s *Service) CreateBinding(ctx context.Context, orgID int64, initiatorPod, targetPod string, scopes []string) (*channel.PodBinding, error) {
	binding := &channel.PodBinding{
		OrganizationID: orgID,
		InitiatorPod:   initiatorPod,
		TargetPod:      targetPod,
		GrantedScopes:    scopes,
		Status:           channel.BindingStatusPending,
	}

	if err := s.db.WithContext(ctx).Create(binding).Error; err != nil {
		return nil, err
	}

	return binding, nil
}

// GetBinding returns a binding by ID
func (s *Service) GetBinding(ctx context.Context, bindingID int64) (*channel.PodBinding, error) {
	var binding channel.PodBinding
	if err := s.db.WithContext(ctx).First(&binding, bindingID).Error; err != nil {
		return nil, err
	}
	return &binding, nil
}

// GetBindingByPods returns a binding between two pods
func (s *Service) GetBindingByPods(ctx context.Context, initiator, target string) (*channel.PodBinding, error) {
	var binding channel.PodBinding
	if err := s.db.WithContext(ctx).
		Where("initiator_pod = ? AND target_pod = ?", initiator, target).
		First(&binding).Error; err != nil {
		return nil, err
	}
	return &binding, nil
}

// ListBindingsForPod returns all bindings for a pod (as initiator or target)
func (s *Service) ListBindingsForPod(ctx context.Context, podKey string) ([]*channel.PodBinding, error) {
	var bindings []*channel.PodBinding
	if err := s.db.WithContext(ctx).
		Where("initiator_pod = ? OR target_pod = ?", podKey, podKey).
		Find(&bindings).Error; err != nil {
		return nil, err
	}
	return bindings, nil
}

// ApproveBinding approves a binding request
func (s *Service) ApproveBinding(ctx context.Context, bindingID int64, scopes []string) error {
	return s.db.WithContext(ctx).Model(&channel.PodBinding{}).
		Where("id = ?", bindingID).
		Updates(map[string]interface{}{
			"status":         channel.BindingStatusApproved,
			"granted_scopes": scopes,
		}).Error
}

// RejectBinding rejects a binding request
func (s *Service) RejectBinding(ctx context.Context, bindingID int64) error {
	return s.db.WithContext(ctx).Model(&channel.PodBinding{}).
		Where("id = ?", bindingID).
		Update("status", channel.BindingStatusRejected).Error
}

// RevokeBinding revokes an approved binding
func (s *Service) RevokeBinding(ctx context.Context, bindingID int64) error {
	return s.db.WithContext(ctx).Model(&channel.PodBinding{}).
		Where("id = ?", bindingID).
		Update("status", channel.BindingStatusRevoked).Error
}

// ChannelPod represents a pod joined to a channel
type ChannelPod struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	ChannelID int64     `gorm:"not null;index" json:"channel_id"`
	PodKey    string    `gorm:"size:100;not null" json:"pod_key"`
	JoinedAt  time.Time `gorm:"not null;default:now()" json:"joined_at"`
}

func (ChannelPod) TableName() string {
	return "channel_pods"
}

// JoinChannel adds a pod to a channel
func (s *Service) JoinChannel(ctx context.Context, channelID int64, podKey string) error {
	cp := &ChannelPod{
		ChannelID: channelID,
		PodKey:    podKey,
		JoinedAt:  time.Now(),
	}
	return s.db.WithContext(ctx).Create(cp).Error
}

// LeaveChannel removes a pod from a channel
func (s *Service) LeaveChannel(ctx context.Context, channelID int64, podKey string) error {
	return s.db.WithContext(ctx).
		Where("channel_id = ? AND pod_key = ?", channelID, podKey).
		Delete(&ChannelPod{}).Error
}

// GetChannelPods returns pods joined to a channel
func (s *Service) GetChannelPods(ctx context.Context, channelID int64) ([]*agentpod.Pod, error) {
	var channelPods []ChannelPod
	if err := s.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Find(&channelPods).Error; err != nil {
		return nil, err
	}

	if len(channelPods) == 0 {
		return []*agentpod.Pod{}, nil
	}

	podKeys := make([]string, len(channelPods))
	for i, cp := range channelPods {
		podKeys[i] = cp.PodKey
	}

	var pods []*agentpod.Pod
	if err := s.db.WithContext(ctx).
		Where("pod_key IN ?", podKeys).
		Find(&pods).Error; err != nil {
		return nil, err
	}

	return pods, nil
}

// ========== Enhanced Message Service ==========

// SendSystemMessage sends a system message to a channel
func (s *Service) SendSystemMessage(ctx context.Context, channelID int64, content string) (*channel.Message, error) {
	return s.SendMessage(ctx, channelID, nil, nil, channel.MessageTypeSystem, content, channel.MessageMetadata{})
}

// SendMessageAsUser sends a message as a user (human) to a channel
func (s *Service) SendMessageAsUser(ctx context.Context, channelID int64, userID int64, content string, metadata channel.MessageMetadata) (*channel.Message, error) {
	return s.SendMessage(ctx, channelID, nil, &userID, channel.MessageTypeText, content, metadata)
}

// SendMessageAsPod sends a message as a pod (agent) to a channel
func (s *Service) SendMessageAsPod(ctx context.Context, channelID int64, podKey string, content string, metadata channel.MessageMetadata) (*channel.Message, error) {
	return s.SendMessage(ctx, channelID, &podKey, nil, channel.MessageTypeText, content, metadata)
}

// GetMessagesMentioning returns messages mentioning a specific pod
func (s *Service) GetMessagesMentioning(ctx context.Context, channelID int64, podKey string, limit int) ([]*channel.Message, error) {
	var messages []*channel.Message
	// Search in content for @pod_key mentions
	pattern := "@" + podKey
	if err := s.db.WithContext(ctx).
		Where("channel_id = ? AND content LIKE ?", channelID, "%"+pattern+"%").
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// GetRecentMessages returns the most recent messages from a channel
func (s *Service) GetRecentMessages(ctx context.Context, channelID int64, limit int) ([]*channel.Message, error) {
	var messages []*channel.Message
	if err := s.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// ========== Access Tracking (Alternative to Explicit Join) ==========

// ChannelAccess tracks pod access to a channel
type ChannelAccess struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	ChannelID  int64     `gorm:"not null;index" json:"channel_id"`
	PodKey     *string   `gorm:"size:100;index" json:"pod_key,omitempty"`
	UserID     *int64    `gorm:"index" json:"user_id,omitempty"`
	LastAccess time.Time `gorm:"not null;default:now()" json:"last_access"`
}

func (ChannelAccess) TableName() string {
	return "channel_access"
}

// TrackAccess records a pod or user accessing a channel
func (s *Service) TrackAccess(ctx context.Context, channelID int64, podKey *string, userID *int64) error {
	// Upsert: update if exists, create if not
	access := &ChannelAccess{
		ChannelID:  channelID,
		PodKey:     podKey,
		UserID:     userID,
		LastAccess: time.Now(),
	}

	// Try to find existing
	query := s.db.WithContext(ctx).Where("channel_id = ?", channelID)
	if podKey != nil {
		query = query.Where("pod_key = ?", *podKey)
	}
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	var existing ChannelAccess
	if err := query.First(&existing).Error; err == nil {
		// Update last access
		return s.db.WithContext(ctx).Model(&existing).Update("last_access", time.Now()).Error
	}

	// Create new
	return s.db.WithContext(ctx).Create(access).Error
}

// GetChannelsForPod returns channels a pod has accessed
func (s *Service) GetChannelsForPod(ctx context.Context, podKey string) ([]*channel.Channel, error) {
	var accesses []ChannelAccess
	if err := s.db.WithContext(ctx).
		Where("pod_key = ?", podKey).
		Find(&accesses).Error; err != nil {
		return nil, err
	}

	if len(accesses) == 0 {
		return []*channel.Channel{}, nil
	}

	channelIDs := make([]int64, len(accesses))
	for i, a := range accesses {
		channelIDs[i] = a.ChannelID
	}

	var channels []*channel.Channel
	if err := s.db.WithContext(ctx).
		Where("id IN ?", channelIDs).
		Find(&channels).Error; err != nil {
		return nil, err
	}

	return channels, nil
}

// HasAccessed checks if a pod has accessed a channel
func (s *Service) HasAccessed(ctx context.Context, channelID int64, podKey string) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&ChannelAccess{}).
		Where("channel_id = ? AND pod_key = ?", channelID, podKey).
		Count(&count).Error
	return count > 0, err
}

// GetAccessCount returns the number of unique accessors for a channel
func (s *Service) GetAccessCount(ctx context.Context, channelID int64) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&ChannelAccess{}).
		Where("channel_id = ?", channelID).
		Count(&count).Error
	return count, err
}
