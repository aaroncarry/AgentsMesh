package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/anthropics/agentmesh/runner/internal/client"
	"github.com/anthropics/agentmesh/runner/internal/terminal"
)

// PodHandler handles pod-related commands from the server.
// Implements the Strategy pattern for different command handlers.
type PodHandler struct {
	runner      *Runner
	termManager *terminal.Manager
	eventSender EventSender
	podStore    PodStore
}

// EventSender sends events to the server.
type EventSender interface {
	SendPodStatus(podKey, status string, data map[string]interface{})
	SendTerminalOutput(podKey string, data []byte) error
}

// PodStore manages pod state.
type PodStore interface {
	Get(podKey string) (*Pod, bool)
	Put(podKey string, pod *Pod)
	Delete(podKey string) *Pod
	Count() int
	All() []*Pod
}

// NewPodHandler creates a new pod handler.
func NewPodHandler(runner *Runner, termManager *terminal.Manager, eventSender EventSender, store PodStore) *PodHandler {
	return &PodHandler{
		runner:      runner,
		termManager: termManager,
		eventSender: eventSender,
		podStore:    store,
	}
}

// HandleMessage routes a message to the appropriate handler.
func (h *PodHandler) HandleMessage(ctx context.Context, msg *client.Message) error {
	switch msg.Type {
	case client.MessageTypePodStart:
		return h.handlePodStart(ctx, msg)
	case client.MessageTypePodStop:
		return h.handlePodStop(ctx, msg)
	case client.MessageTypeTerminalInput:
		return h.handleTerminalInput(ctx, msg)
	case client.MessageTypeTerminalResize:
		return h.handleTerminalResize(ctx, msg)
	case client.MessageTypePodList:
		return h.handlePodList(ctx, msg)
	default:
		log.Printf("[pod_handler] Unknown message type: %s", msg.Type)
		return nil
	}
}

// handlePodStart handles pod start requests.
func (h *PodHandler) handlePodStart(ctx context.Context, msg *client.Message) error {
	var payload PodStartPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid pod start payload: %w", err)
	}

	log.Printf("[pod_handler] Starting pod: pod_key=%s, agent=%s",
		payload.PodKey, payload.AgentType)

	// Check capacity
	if h.runner.cfg.MaxConcurrentPods > 0 && h.podStore.Count() >= h.runner.cfg.MaxConcurrentPods {
		h.sendPodError(payload.PodKey, "max concurrent pods reached")
		return fmt.Errorf("max concurrent pods reached")
	}

	// Build the pod using PodBuilder
	builder := NewPodBuilder(h.runner).
		WithPodKey(payload.PodKey).
		WithAgentType(payload.AgentType).
		WithLaunchCommand(payload.LaunchCommand, payload.LaunchArgs).
		WithEnvVars(payload.EnvVars).
		WithTerminalSize(payload.Rows, payload.Cols).
		WithInitialPrompt(payload.InitialPrompt)

	// Configure repository if specified
	if payload.RepositoryURL != "" {
		builder.WithRepository(payload.RepositoryURL, payload.Branch)
	}

	// Configure worktree if ticket identifier is specified
	if payload.TicketIdentifier != "" {
		builder.WithWorktree(payload.TicketIdentifier)
	}

	// Enable sandbox mode if sandbox manager is available
	// This enables the new plugin-based environment setup
	if h.runner.sandboxManager != nil {
		pluginConfig := payload.ToPluginConfig()
		builder.WithSandbox(pluginConfig)
	}

	// Build and start the pod
	pod, err := builder.Build(ctx)
	if err != nil {
		h.sendPodError(payload.PodKey, fmt.Sprintf("failed to build pod: %v", err))
		return fmt.Errorf("failed to build pod: %w", err)
	}

	// Set up output handler
	pod.OnOutput = func(data []byte) {
		if err := h.eventSender.SendTerminalOutput(payload.PodKey, data); err != nil {
			log.Printf("[pod_handler] Failed to send terminal output: %v", err)
		}
	}

	// Set up exit handler
	pod.OnExit = func(exitCode int) {
		h.handlePodExit(payload.PodKey, exitCode)
	}

	// Start the terminal
	if err := pod.Terminal.Start(); err != nil {
		h.sendPodError(payload.PodKey, fmt.Sprintf("failed to start terminal: %v", err))
		return fmt.Errorf("failed to start terminal: %w", err)
	}

	// Store the pod
	h.podStore.Put(payload.PodKey, pod)

	// Send initial prompt if specified
	if payload.InitialPrompt != "" {
		time.AfterFunc(500*time.Millisecond, func() {
			if err := pod.Terminal.Write([]byte(payload.InitialPrompt + "\n")); err != nil {
				log.Printf("[pod_handler] Failed to send initial prompt: %v", err)
			}
		})
	}

	// Notify server
	h.eventSender.SendPodStatus(payload.PodKey, "started", map[string]interface{}{
		"pid": pod.Terminal.PID(),
	})

	log.Printf("[pod_handler] Pod started: pod_key=%s, pid=%d",
		payload.PodKey, pod.Terminal.PID())

	return nil
}

// handlePodStop handles pod stop requests.
func (h *PodHandler) handlePodStop(ctx context.Context, msg *client.Message) error {
	var payload PodStopPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid pod stop payload: %w", err)
	}

	log.Printf("[pod_handler] Stopping pod: pod_key=%s", payload.PodKey)

	pod := h.podStore.Delete(payload.PodKey)
	if pod == nil {
		return fmt.Errorf("pod not found: %s", payload.PodKey)
	}

	// Stop terminal
	if pod.Terminal != nil {
		pod.Terminal.Stop()
	}

	// Clean up sandbox if sandbox manager is available
	if h.runner.sandboxManager != nil {
		if err := h.runner.sandboxManager.Cleanup(payload.PodKey); err != nil {
			log.Printf("[pod_handler] Warning: failed to cleanup sandbox: %v", err)
		}
	} else if pod.WorktreePath != "" && h.runner.workspace != nil {
		// Legacy mode: clean up worktree directly
		if err := h.runner.workspace.RemoveWorktree(ctx, pod.WorktreePath); err != nil {
			log.Printf("[pod_handler] Warning: failed to remove worktree: %v", err)
		}
	}

	// Notify server
	h.eventSender.SendPodStatus(payload.PodKey, "stopped", nil)

	log.Printf("[pod_handler] Pod stopped: pod_key=%s", payload.PodKey)
	return nil
}

// handleTerminalInput handles terminal input.
func (h *PodHandler) handleTerminalInput(ctx context.Context, msg *client.Message) error {
	var payload TerminalInputPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid terminal input payload: %w", err)
	}

	pod, ok := h.podStore.Get(payload.PodKey)
	if !ok {
		return fmt.Errorf("pod not found: %s", payload.PodKey)
	}

	return pod.Terminal.Write(payload.Data)
}

// handleTerminalResize handles terminal resize requests.
func (h *PodHandler) handleTerminalResize(ctx context.Context, msg *client.Message) error {
	var payload TerminalResizePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid terminal resize payload: %w", err)
	}

	pod, ok := h.podStore.Get(payload.PodKey)
	if !ok {
		return fmt.Errorf("pod not found: %s", payload.PodKey)
	}

	return pod.Terminal.Resize(payload.Rows, payload.Cols)
}

// PodListPayload represents the payload for pod list request.
type PodListPayload struct {
	RequestID string `json:"request_id"`
}

// handlePodList handles pod list requests.
func (h *PodHandler) handlePodList(ctx context.Context, msg *client.Message) error {
	var payload PodListPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid pod list payload: %w", err)
	}

	pods := h.podStore.All()
	podInfos := make([]map[string]interface{}, 0, len(pods))

	for _, p := range pods {
		info := map[string]interface{}{
			"pod_key":        p.PodKey,
			"agent_type":     p.AgentType,
			"status":         p.Status,
			"started_at":     p.StartedAt.Format(time.RFC3339),
			"worktree_path":  p.WorktreePath,
			"repository_url": p.RepositoryURL,
		}
		if p.Terminal != nil {
			info["pid"] = p.Terminal.PID()
		}
		podInfos = append(podInfos, info)
	}

	h.eventSender.SendPodStatus("", "pod_list", map[string]interface{}{
		"request_id": payload.RequestID,
		"pods":       podInfos,
	})

	return nil
}

// handlePodExit handles terminal exit events.
func (h *PodHandler) handlePodExit(podKey string, exitCode int) {
	log.Printf("[pod_handler] Pod exited: pod_key=%s, exit_code=%d",
		podKey, exitCode)

	pod := h.podStore.Delete(podKey)
	if pod != nil {
		pod.Status = PodStatusStopped
	}

	// Clean up sandbox if sandbox manager is available
	if h.runner.sandboxManager != nil {
		if err := h.runner.sandboxManager.Cleanup(podKey); err != nil {
			log.Printf("[pod_handler] Warning: failed to cleanup sandbox on exit: %v", err)
		}
	}

	h.eventSender.SendPodStatus(podKey, "exited", map[string]interface{}{
		"exit_code": exitCode,
	})
}

// sendPodError sends a pod error notification.
func (h *PodHandler) sendPodError(podKey, errorMsg string) {
	h.eventSender.SendPodStatus(podKey, "error", map[string]interface{}{
		"error": errorMsg,
	})
}

// InMemoryPodStore is a simple in-memory pod store.
type InMemoryPodStore struct {
	pods map[string]*Pod
	mu   sync.RWMutex
}

// NewInMemoryPodStore creates a new in-memory pod store.
func NewInMemoryPodStore() *InMemoryPodStore {
	return &InMemoryPodStore{
		pods: make(map[string]*Pod),
	}
}

// Get retrieves a pod by key.
func (s *InMemoryPodStore) Get(podKey string) (*Pod, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pod, ok := s.pods[podKey]
	return pod, ok
}

// Put stores a pod.
func (s *InMemoryPodStore) Put(podKey string, pod *Pod) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pods[podKey] = pod
}

// Delete removes and returns a pod.
func (s *InMemoryPodStore) Delete(podKey string) *Pod {
	s.mu.Lock()
	defer s.mu.Unlock()
	pod, ok := s.pods[podKey]
	if ok {
		delete(s.pods, podKey)
	}
	return pod
}

// Count returns the number of pods.
func (s *InMemoryPodStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.pods)
}

// All returns all pods.
func (s *InMemoryPodStore) All() []*Pod {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pods := make([]*Pod, 0, len(s.pods))
	for _, pod := range s.pods {
		pods = append(pods, pod)
	}
	return pods
}

