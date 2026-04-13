package channel

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
)

// MemberIDProvider resolves member user IDs for a channel.
// Used by the event hook to attach TargetUserIDs for member-only delivery.
type MemberIDProvider interface {
	GetMemberUserIDs(ctx context.Context, channelID int64) ([]int64, error)
}

// NewEventPublishHook creates a hook that publishes channel:message events.
// members is used to resolve channel member IDs for targeted WebSocket delivery.
func NewEventPublishHook(eb *eventbus.EventBus, userNames UserNameResolver, members MemberIDProvider) PostSendHook {
	return func(ctx context.Context, mc *MessageContext) error {
		if eb == nil {
			return nil
		}

		msgData := eventbus.ChannelMessageData{
			ID:           mc.Message.ID,
			ChannelID:    mc.Message.ChannelID,
			SenderPod:    mc.Message.SenderPod,
			SenderUserID: mc.Message.SenderUserID,
			SenderName:   resolveSenderName(ctx, mc, userNames),
			MessageType:  mc.Message.MessageType,
			Content:      mc.Message.Content,
			Metadata:     mc.Message.Metadata,
			CreatedAt:    mc.Message.CreatedAt.Format(time.RFC3339),
		}
		data, err := json.Marshal(msgData)
		if err != nil {
			slog.Error("failed to marshal channel message event", "error", err)
			return err
		}

		var targetUserIDs []int64
		if members != nil {
			targetUserIDs, _ = members.GetMemberUserIDs(ctx, mc.Channel.ID)
		}

		eb.Publish(ctx, &eventbus.Event{
			Type:           eventbus.EventChannelMessage,
			Category:       eventbus.CategoryEntity,
			OrganizationID: mc.Channel.OrganizationID,
			EntityType:     "channel",
			EntityID:       strconv.FormatInt(mc.Message.ChannelID, 10),
			Data:           data,
			TargetUserIDs:  targetUserIDs,
		})

		return nil
	}
}
