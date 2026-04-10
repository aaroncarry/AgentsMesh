package infra

import (
	"context"
	"encoding/json"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/channel"
	"gorm.io/gorm"
)

func (r *channelRepository) CreateMessage(ctx context.Context, msg *channel.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *channelRepository) TouchChannel(ctx context.Context, channelID int64) error {
	return r.db.WithContext(ctx).Model(&channel.Channel{}).
		Where("id = ?", channelID).
		Update("updated_at", time.Now()).Error
}

func (r *channelRepository) GetMessages(ctx context.Context, channelID int64, before *time.Time, after *time.Time, limit int) ([]*channel.Message, error) {
	query := r.db.WithContext(ctx).Where("channel_id = ? AND is_deleted = FALSE", channelID)
	if before != nil {
		query = query.Where("created_at < ?", *before)
	}
	if after != nil {
		query = query.Where("created_at > ?", *after)
	}
	// After-only: return oldest-first so hasMore means "newer messages exist beyond the limit".
	// All other cases: newest-first so hasMore means "older messages exist" (load-more / scroll-up).
	order := "created_at DESC, id DESC"
	if after != nil && before == nil {
		order = "created_at ASC, id ASC"
	}
	var messages []*channel.Message
	if err := query.
		Preload("SenderUser").
		Preload("SenderPodInfo").
		Preload("SenderPodInfo.Agent").
		Order(order).
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *channelRepository) GetMessagesMentioning(ctx context.Context, channelID int64, podKey string, limit int) ([]*channel.Message, bool, error) {
	var messages []*channel.Message
	podValuePattern := `%"` + podKey + `"%`
	if err := r.db.WithContext(ctx).
		Where(`channel_id = ? AND is_deleted = FALSE AND CAST(metadata AS TEXT) LIKE '%mentioned_pods%' AND CAST(metadata AS TEXT) LIKE ?`,
			channelID, podValuePattern).
		Order("created_at DESC, id DESC").
		Limit(limit + 1).
		Find(&messages).Error; err != nil {
		return nil, false, err
	}
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}
	return messages, hasMore, nil
}

func (r *channelRepository) GetRecentMessages(ctx context.Context, channelID int64, limit int) ([]*channel.Message, error) {
	var messages []*channel.Message
	if err := r.db.WithContext(ctx).
		Where("channel_id = ? AND is_deleted = FALSE", channelID).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *channelRepository) GetMessageByID(ctx context.Context, messageID int64) (*channel.Message, error) {
	var msg channel.Message
	if err := r.db.WithContext(ctx).First(&msg, messageID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &msg, nil
}

func (r *channelRepository) UpdateMessageContent(ctx context.Context, messageID int64, content string) error {
	return r.db.WithContext(ctx).
		Model(&channel.Message{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"content":   content,
			"edited_at": time.Now(),
		}).Error
}

func (r *channelRepository) SoftDeleteMessage(ctx context.Context, messageID int64) error {
	return r.db.WithContext(ctx).
		Model(&channel.Message{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"content":    "[deleted]",
		}).Error
}

func (r *channelRepository) UpdateMessageMetadata(ctx context.Context, messageID int64, metadata map[string]interface{}) error {
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Exec(
		`UPDATE channel_messages SET metadata = COALESCE(metadata, '{}'::jsonb) || ?::jsonb WHERE id = ?`,
		string(jsonData), messageID,
	).Error
}

func (r *channelRepository) GetMessagesBefore(ctx context.Context, channelID int64, beforeID int64, limit int) ([]*channel.Message, error) {
	var messages []*channel.Message
	query := r.db.WithContext(ctx).
		Where("channel_id = ? AND id < ? AND is_deleted = FALSE", channelID, beforeID).
		Preload("SenderUser").
		Preload("SenderPodInfo").
		Preload("SenderPodInfo.Agent").
		Order("id DESC").
		Limit(limit)
	if err := query.Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}
