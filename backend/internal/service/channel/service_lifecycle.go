package channel

import (
	"context"
	"log/slog"

	channelDomain "github.com/anthropics/agentsmesh/backend/internal/domain/channel"
)

func (s *Service) ArchiveChannel(ctx context.Context, channelID int64) error {
	if err := s.repo.SetArchived(ctx, channelID, true); err != nil {
		slog.Error("failed to archive channel", "channel_id", channelID, "error", err)
		return err
	}
	slog.Info("channel archived", "channel_id", channelID)
	return nil
}

func (s *Service) UnarchiveChannel(ctx context.Context, channelID int64) error {
	if err := s.repo.SetArchived(ctx, channelID, false); err != nil {
		slog.Error("failed to unarchive channel", "channel_id", channelID, "error", err)
		return err
	}
	slog.Info("channel unarchived", "channel_id", channelID)
	return nil
}

func (s *Service) DeleteChannel(ctx context.Context, channelID int64) error {
	if err := s.repo.DeleteWithCleanup(ctx, channelID); err != nil {
		slog.Error("failed to delete channel", "channel_id", channelID, "error", err)
		return err
	}
	slog.Info("channel deleted", "channel_id", channelID)
	return nil
}

func (s *Service) DeleteChannelsByOrg(ctx context.Context, orgID int64) error {
	if err := s.repo.DeleteChannelsByOrg(ctx, orgID); err != nil {
		slog.Error("failed to delete channels by org", "org_id", orgID, "error", err)
		return err
	}
	slog.Info("channels deleted for org", "org_id", orgID)
	return nil
}

func (s *Service) CleanupUserReferences(ctx context.Context, userID int64) error {
	return s.repo.CleanupUserReferences(ctx, userID)
}

func (s *Service) GetChannelsByTicket(ctx context.Context, ticketID int64) ([]*channelDomain.Channel, error) {
	return s.repo.GetByTicketID(ctx, ticketID)
}
