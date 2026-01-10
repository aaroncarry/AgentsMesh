package devmesh

import (
	"context"
	"errors"

	"github.com/anthropics/agentmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentmesh/backend/internal/domain/channel"
	"github.com/anthropics/agentmesh/backend/internal/domain/devmesh"
	bindingService "github.com/anthropics/agentmesh/backend/internal/service/binding"
	channelService "github.com/anthropics/agentmesh/backend/internal/service/channel"
	podService "github.com/anthropics/agentmesh/backend/internal/service/agentpod"
	"gorm.io/gorm"
)

var (
	ErrTicketNotFound = errors.New("ticket not found")
	ErrRunnerNotFound = errors.New("runner not found")
)

// Service handles DevMesh operations
type Service struct {
	db             *gorm.DB
	podService     *podService.PodService
	channelService *channelService.Service
	bindingService *bindingService.Service
}

// NewService creates a new DevMesh service
func NewService(
	db *gorm.DB,
	ps *podService.PodService,
	cs *channelService.Service,
	bs *bindingService.Service,
) *Service {
	return &Service{
		db:             db,
		podService:     ps,
		channelService: cs,
		bindingService: bs,
	}
}

// GetTopology returns the complete DevMesh topology for an organization
func (s *Service) GetTopology(ctx context.Context, orgID int64) (*devmesh.DevMeshTopology, error) {
	// 1. Get active pods
	pods, _, err := s.podService.ListPods(ctx, orgID, nil, "", 100, 0)
	if err != nil {
		return nil, err
	}

	// Filter to only active pods and convert to nodes
	nodes := make([]devmesh.DevMeshNode, 0)
	podKeys := make([]string, 0)

	for _, pod := range pods {
		if pod.IsActive() {
			node := s.podToNode(pod)
			nodes = append(nodes, node)
			podKeys = append(podKeys, pod.PodKey)
		}
	}

	// 2. Get bindings (edges) for active pods
	edges := make([]devmesh.DevMeshEdge, 0)
	seenBindings := make(map[int64]bool) // Track seen binding IDs to avoid duplicates
	for _, key := range podKeys {
		activeStatus := channel.BindingStatusActive
		bindings, err := s.bindingService.GetBindingsForPod(ctx, key, &activeStatus)
		if err != nil {
			continue
		}
		for _, b := range bindings {
			// Skip if we've already seen this binding (since it appears for both source and target pods)
			if seenBindings[b.ID] {
				continue
			}
			seenBindings[b.ID] = true

			if b.IsActive() {
				edges = append(edges, devmesh.DevMeshEdge{
					ID:            b.ID,
					Source:        b.InitiatorPod,
					Target:        b.TargetPod,
					GrantedScopes: []string(b.GrantedScopes),
					PendingScopes: []string(b.PendingScopes),
					Status:        b.Status,
				})
			}
		}
	}

	// 3. Get channels
	channels, _, err := s.channelService.ListChannels(ctx, orgID, false, 50, 0)
	if err != nil {
		return nil, err
	}

	channelInfos := make([]devmesh.ChannelInfo, 0, len(channels))
	for _, ch := range channels {
		// Get pods in this channel
		channelPods := s.getChannelPods(ctx, ch.ID)

		// Get message count
		messageCount := s.getChannelMessageCount(ctx, ch.ID)

		channelInfos = append(channelInfos, devmesh.ChannelInfo{
			ID:           ch.ID,
			Name:         ch.Name,
			Description:  ch.Description,
			PodKeys:      channelPods,
			MessageCount: messageCount,
			IsArchived:   ch.IsArchived,
		})
	}

	return &devmesh.DevMeshTopology{
		Nodes:    nodes,
		Edges:    edges,
		Channels: channelInfos,
	}, nil
}

// podToNode converts a pod to a DevMesh node
func (s *Service) podToNode(pod *agentpod.Pod) devmesh.DevMeshNode {
	return devmesh.DevMeshNode{
		PodKey:       pod.PodKey,
		Status:       pod.Status,
		AgentStatus:  pod.AgentStatus,
		Model:        pod.Model,
		TicketID:     pod.TicketID,
		RepositoryID: pod.RepositoryID,
		CreatedByID:  pod.CreatedByID,
		RunnerID:     pod.RunnerID,
		StartedAt:    pod.StartedAt,
	}
}

// getChannelPods returns pod keys in a channel
func (s *Service) getChannelPods(ctx context.Context, channelID int64) []string {
	var channelPods []devmesh.ChannelPod
	s.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Find(&channelPods)

	keys := make([]string, len(channelPods))
	for i, cp := range channelPods {
		keys[i] = cp.PodKey
	}
	return keys
}

// getChannelMessageCount returns the message count for a channel
func (s *Service) getChannelMessageCount(ctx context.Context, channelID int64) int {
	var count int64
	s.db.WithContext(ctx).
		Model(&channel.Message{}).
		Where("channel_id = ?", channelID).
		Count(&count)
	return int(count)
}

// CreatePodForTicket creates a new pod associated with a ticket
func (s *Service) CreatePodForTicket(ctx context.Context, req *devmesh.CreatePodForTicketRequest) (*agentpod.Pod, error) {
	return s.podService.CreatePodForTicket(ctx, &podService.CreatePodRequest{
		OrganizationID: req.OrganizationID,
		RunnerID:       req.RunnerID,
		TicketID:       &req.TicketID,
		CreatedByID:    req.CreatedByID,
		InitialPrompt:  req.InitialPrompt,
		Model:          req.Model,
		PermissionMode: req.PermissionMode,
		ThinkLevel:     req.ThinkLevel,
	})
}

// GetPodsForTicket returns all pods associated with a ticket
func (s *Service) GetPodsForTicket(ctx context.Context, ticketID int64) ([]devmesh.DevMeshNode, error) {
	pods, err := s.podService.GetPodsByTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	nodes := make([]devmesh.DevMeshNode, len(pods))
	for i, pod := range pods {
		nodes[i] = s.podToNode(pod)
	}
	return nodes, nil
}

// GetActivePodsForTicket returns only active pods for a ticket
func (s *Service) GetActivePodsForTicket(ctx context.Context, ticketID int64) ([]devmesh.DevMeshNode, error) {
	pods, err := s.podService.GetPodsByTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	nodes := make([]devmesh.DevMeshNode, 0)
	for _, pod := range pods {
		if pod.IsActive() {
			nodes = append(nodes, s.podToNode(pod))
		}
	}
	return nodes, nil
}

// BatchGetTicketPods returns pods for multiple tickets
func (s *Service) BatchGetTicketPods(ctx context.Context, ticketIDs []int64) (*devmesh.BatchTicketPodsResponse, error) {
	// Get all pods for the given ticket IDs
	var pods []*agentpod.Pod
	if err := s.db.WithContext(ctx).
		Where("ticket_id IN ?", ticketIDs).
		Find(&pods).Error; err != nil {
		return nil, err
	}

	// Group by ticket ID
	result := make(map[int64][]devmesh.DevMeshNode)
	for _, pod := range pods {
		if pod.TicketID != nil {
			ticketID := *pod.TicketID
			if _, exists := result[ticketID]; !exists {
				result[ticketID] = make([]devmesh.DevMeshNode, 0)
			}
			result[ticketID] = append(result[ticketID], s.podToNode(pod))
		}
	}

	// Ensure all requested ticket IDs are in the result (even if empty)
	for _, id := range ticketIDs {
		if _, exists := result[id]; !exists {
			result[id] = make([]devmesh.DevMeshNode, 0)
		}
	}

	return &devmesh.BatchTicketPodsResponse{
		TicketPods: result,
	}, nil
}

// JoinChannel adds a pod to a channel
func (s *Service) JoinChannel(ctx context.Context, channelID int64, podKey string) error {
	cp := &devmesh.ChannelPod{
		ChannelID: channelID,
		PodKey:    podKey,
	}
	return s.db.WithContext(ctx).Create(cp).Error
}

// LeaveChannel removes a pod from a channel
func (s *Service) LeaveChannel(ctx context.Context, channelID int64, podKey string) error {
	return s.db.WithContext(ctx).
		Where("channel_id = ? AND pod_key = ?", channelID, podKey).
		Delete(&devmesh.ChannelPod{}).Error
}

// RecordChannelAccess records access to a channel
func (s *Service) RecordChannelAccess(ctx context.Context, channelID int64, podKey *string, userID *int64) error {
	access := &devmesh.ChannelAccess{
		ChannelID: channelID,
		PodKey:    podKey,
		UserID:    userID,
	}
	return s.db.WithContext(ctx).Create(access).Error
}
