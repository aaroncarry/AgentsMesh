package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/channel"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
)

// EditMessage edits a message's content. Only the original sender can edit.
func (s *Service) EditMessage(ctx context.Context, channelID, messageID, senderUserID int64, newContent string) (*channel.Message, error) {
	ch, err := s.GetChannel(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if ch.IsArchived {
		return nil, ErrChannelArchived
	}

	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if msg == nil || msg.ChannelID != channelID {
		return nil, ErrMessageNotFound
	}
	if msg.SenderUserID == nil || *msg.SenderUserID != senderUserID {
		return nil, ErrNotMessageSender
	}

	if err := s.repo.UpdateMessageContent(ctx, messageID, newContent); err != nil {
		return nil, err
	}

	s.publishMessageEvent(ch.OrganizationID, eventbus.EventChannelMessageEdited, map[string]interface{}{
		"channel_id": channelID,
		"id":         messageID,
		"content":    newContent,
		"edited_at":  time.Now().Format(time.RFC3339),
	})

	return s.repo.GetMessageByID(ctx, messageID)
}

// DeleteMessage soft-deletes a message. Only the original sender can delete.
func (s *Service) DeleteMessage(ctx context.Context, channelID, messageID, senderUserID int64) error {
	ch, err := s.GetChannel(ctx, channelID)
	if err != nil {
		return err
	}
	if ch.IsArchived {
		return ErrChannelArchived
	}

	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return err
	}
	if msg == nil || msg.ChannelID != channelID {
		return ErrMessageNotFound
	}
	if msg.SenderUserID == nil || *msg.SenderUserID != senderUserID {
		return ErrNotMessageSender
	}

	if err := s.repo.SoftDeleteMessage(ctx, messageID); err != nil {
		return err
	}

	s.publishMessageEvent(ch.OrganizationID, eventbus.EventChannelMessageDeleted, map[string]interface{}{
		"channel_id": channelID,
		"id":         messageID,
	})

	return nil
}

func (s *Service) publishMessageEvent(orgID int64, eventType eventbus.EventType, data map[string]interface{}) {
	if s.eventBus == nil {
		return
	}
	payload, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal message event", "error", err)
		return
	}
	channelID, _ := data["channel_id"].(int64)
	ctx := context.Background()
	s.eventBus.Publish(ctx, &eventbus.Event{
		Type:           eventType,
		Category:       eventbus.CategoryEntity,
		OrganizationID: orgID,
		EntityType:     "channel",
		EntityID:       fmt.Sprintf("%d", channelID),
		Data:           payload,
		Timestamp:      time.Now().UnixMilli(),
	})
}
