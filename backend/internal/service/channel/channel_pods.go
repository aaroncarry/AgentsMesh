package channel

import (
	"context"
	"log/slog"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	channelDomain "github.com/anthropics/agentsmesh/backend/internal/domain/channel"
)

// PodCreatorResolver resolves the human creator of a pod.
// When a pod is created by another pod, this traces back to the original human user.
type PodCreatorResolver interface {
	GetPodByKey(ctx context.Context, podKey string) (*agentpod.Pod, error)
}

// SetPodCreatorResolver injects the pod creator resolution capability.
// Called during wiring in main.go to avoid import cycles.
func (s *Service) SetPodCreatorResolver(resolver PodCreatorResolver) {
	s.podCreatorResolver = resolver
}

// JoinChannel adds a pod to a channel and auto-joins the pod's human creator as a member.
func (s *Service) JoinChannel(ctx context.Context, channelID int64, podKey string) error {
	if err := s.repo.AddPodToChannel(ctx, channelID, podKey); err != nil {
		return err
	}
	s.ensurePodCreatorIsMember(ctx, channelID, podKey)
	return nil
}

// LeaveChannel removes a pod from a channel
func (s *Service) LeaveChannel(ctx context.Context, channelID int64, podKey string) error {
	return s.repo.RemovePodFromChannel(ctx, channelID, podKey)
}

// GetChannelPods returns pods joined to a channel
func (s *Service) GetChannelPods(ctx context.Context, channelID int64) ([]*agentpod.Pod, error) {
	return s.repo.GetChannelPods(ctx, channelID)
}

// ensurePodCreatorIsMember resolves a pod's human creator and adds them as a channel member.
func (s *Service) ensurePodCreatorIsMember(ctx context.Context, channelID int64, podKey string) {
	if s.podCreatorResolver == nil {
		return
	}
	pod, err := s.podCreatorResolver.GetPodByKey(ctx, podKey)
	if err != nil || pod == nil {
		slog.WarnContext(ctx, "failed to resolve pod creator for channel membership", "pod_key", podKey, "error", err)
		return
	}
	_ = s.repo.AddMemberWithRole(ctx, channelID, pod.CreatedByID, channelDomain.RoleMember)
}
