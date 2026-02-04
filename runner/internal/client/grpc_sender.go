// Package client provides gRPC connection management for Runner.
package client

import (
	"fmt"
	"time"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

// sendControl queues a control message (high priority).
// Control messages include: heartbeat, pod_created, pod_terminated, pty_resized, error.
// These messages should never be blocked by terminal output.
// Returns error if connection is closed, stopped, or channel is full.
func (c *GRPCConnection) sendControl(msg *runnerv1.RunnerMessage) error {
	c.mu.Lock()
	if c.stream == nil {
		c.mu.Unlock()
		return fmt.Errorf("stream not connected")
	}
	c.mu.Unlock()

	select {
	case c.controlCh <- msg:
		return nil
	case <-c.stopCh:
		return fmt.Errorf("connection stopped")
	default:
		return fmt.Errorf("control buffer full")
	}
}

// sendTerminal queues a terminal message (low priority).
// Terminal messages include: agent_status.
// NOTE: terminal_output removed - terminal output is exclusively streamed via Relay.
// These messages are dropped silently if buffer is full.
// Returns nil even when dropped to avoid blocking callers.
//
// IMPORTANT: Messages are rejected before initialization completes.
// This prevents queue buildup during reconnection handshake, which could
// cause gRPC flow control to block the initialize_result response.
func (c *GRPCConnection) sendTerminal(msg *runnerv1.RunnerMessage) error {
	c.mu.Lock()
	stream := c.stream
	initialized := c.initialized
	c.mu.Unlock()

	// Reject messages before initialization completes
	// During reconnection, old Pods may still produce output, but sending it
	// before handshake completes can block the gRPC stream and cause deadlock
	if !initialized {
		logger.Terminal().Debug("sendTerminal: not initialized, dropping message")
		return nil // Silent drop, not an error
	}

	if stream == nil {
		logger.Terminal().Debug("sendTerminal: stream not connected")
		return fmt.Errorf("stream not connected")
	}

	select {
	case c.terminalCh <- msg:
		logger.Terminal().Debug("sendTerminal: message queued",
			"queue_len", len(c.terminalCh))
		return nil
	case <-c.stopCh:
		logger.Terminal().Debug("sendTerminal: connection stopped")
		return fmt.Errorf("connection stopped")
	default:
		// TUI frames are expendable - drop silently
		logger.GRPC().Debug("Terminal output dropped (queue full)",
			"queue_usage", c.QueueUsage())
		return nil
	}
}

// SendPodCreated sends a pod_created event to the server (control message).
func (c *GRPCConnection) SendPodCreated(podKey string, pid int32, sandboxPath, branchName string) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_PodCreated{
			PodCreated: &runnerv1.PodCreatedEvent{
				PodKey:      podKey,
				Pid:         pid,
				SandboxPath: sandboxPath,
				BranchName:  branchName,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	return c.sendControl(msg)
}

// SendPodTerminated sends a pod_terminated event to the server (control message).
func (c *GRPCConnection) SendPodTerminated(podKey string, exitCode int32, errorMsg string) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_PodTerminated{
			PodTerminated: &runnerv1.PodTerminatedEvent{
				PodKey:       podKey,
				ExitCode:     exitCode,
				ErrorMessage: errorMsg,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	return c.sendControl(msg)
}

// NOTE: SendTerminalOutput removed - terminal output is exclusively streamed via Relay

// SendAgentStatus sends an agent status change event to the server (terminal message).
func (c *GRPCConnection) SendAgentStatus(podKey string, status string) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_AgentStatus{
			AgentStatus: &runnerv1.AgentStatusEvent{
				PodKey: podKey,
				Status: status,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	return c.sendTerminal(msg)
}

// SendPtyResized sends a PTY resize event to the server (control message).
func (c *GRPCConnection) SendPtyResized(podKey string, cols, rows int32) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_PtyResized{
			PtyResized: &runnerv1.PtyResizedEvent{
				PodKey: podKey,
				Cols:   cols,
				Rows:   rows,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	return c.sendControl(msg)
}

// SendError sends an error event to the server (control message).
func (c *GRPCConnection) SendError(podKey, code, message string) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_Error{
			Error: &runnerv1.ErrorEvent{
				PodKey:  podKey,
				Code:    code,
				Message: message,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	return c.sendControl(msg)
}

// SendPodInitProgress sends a pod initialization progress event to the server (control message).
func (c *GRPCConnection) SendPodInitProgress(podKey, phase string, progress int32, message string) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_PodInitProgress{
			PodInitProgress: &runnerv1.PodInitProgressEvent{
				PodKey:   podKey,
				Phase:    phase,
				Progress: progress,
				Message:  message,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	return c.sendControl(msg)
}

// SendRequestRelayToken sends a request for a new relay token to the server.
// This is called when the relay connection fails due to token expiration.
func (c *GRPCConnection) SendRequestRelayToken(podKey, sessionID, relayURL string) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_RequestRelayToken{
			RequestRelayToken: &runnerv1.RequestRelayTokenEvent{
				PodKey:    podKey,
				SessionId: sessionID,
				RelayUrl:  relayURL,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	return c.sendControl(msg)
}

// SendOSCNotification sends an OSC notification event to the server (control message).
// This is triggered by OSC 777 (iTerm2/Kitty) or OSC 9 (ConEmu/Windows Terminal) sequences.
// Uses controlCh for high priority delivery (not affected by terminal output throttling).
func (c *GRPCConnection) SendOSCNotification(podKey, title, body string) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_OscNotification{
			OscNotification: &runnerv1.OSCNotificationEvent{
				PodKey:    podKey,
				Title:     title,
				Body:      body,
				Timestamp: time.Now().UnixMilli(),
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	return c.sendControl(msg)
}

// SendOSCTitle sends an OSC title change event to the server (control message).
// This is triggered by OSC 0/2 sequences for window/tab title changes.
func (c *GRPCConnection) SendOSCTitle(podKey, title string) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_OscTitle{
			OscTitle: &runnerv1.OSCTitleEvent{
				PodKey: podKey,
				Title:  title,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	return c.sendControl(msg)
}

// SendMessage sends a raw RunnerMessage to the server.
// Used for Autopilot events and other custom messages.
func (c *GRPCConnection) SendMessage(msg *runnerv1.RunnerMessage) error {
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixMilli()
	}
	return c.sendControl(msg)
}

// SendSandboxesStatus sends sandbox status query response to the server (control message).
func (c *GRPCConnection) SendSandboxesStatus(requestID string, sandboxes []*SandboxStatusInfo) error {
	// Convert SandboxStatusInfo to proto
	protoSandboxes := make([]*runnerv1.SandboxStatus, len(sandboxes))
	for i, s := range sandboxes {
		protoSandboxes[i] = &runnerv1.SandboxStatus{
			PodKey:                s.PodKey,
			Exists:                s.Exists,
			SandboxPath:           s.SandboxPath,
			RepositoryUrl:         s.RepositoryURL,
			BranchName:            s.BranchName,
			CurrentCommit:         s.CurrentCommit,
			SizeBytes:             s.SizeBytes,
			LastModified:          s.LastModified,
			HasUncommittedChanges: s.HasUncommittedChanges,
			CanResume:             s.CanResume,
			Error:                 s.Error,
		}
	}

	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_SandboxesStatus{
			SandboxesStatus: &runnerv1.SandboxesStatusEvent{
				RequestId: requestID,
				Sandboxes: protoSandboxes,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	return c.sendControl(msg)
}

// sendError sends an error event back to the server (internal use, control message).
func (c *GRPCConnection) sendError(podKey, code, message string) {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_Error{
			Error: &runnerv1.ErrorEvent{
				PodKey:  podKey,
				Code:    code,
				Message: message,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	if err := c.sendControl(msg); err != nil {
		logger.GRPC().Error("Failed to send error", "error", err)
	}
}

// QueueLength returns the current terminal send queue length.
func (c *GRPCConnection) QueueLength() int {
	return len(c.terminalCh)
}

// QueueCapacity returns the terminal send queue capacity.
func (c *GRPCConnection) QueueCapacity() int {
	return cap(c.terminalCh)
}

// QueueUsage returns the terminal queue usage ratio (0.0 to 1.0).
// Used for monitoring queue pressure.
func (c *GRPCConnection) QueueUsage() float64 {
	return float64(len(c.terminalCh)) / float64(cap(c.terminalCh))
}

// drainTerminalQueue clears all pending messages in the terminal queue.
// Called before reconnection to discard stale terminal output.
// TUI frames are expendable - old frames are irrelevant after reconnection.
func (c *GRPCConnection) drainTerminalQueue() {
	drained := 0
	for {
		select {
		case <-c.terminalCh:
			drained++
		default:
			if drained > 0 {
				logger.GRPC().Info("Drained stale terminal queue before reconnection",
					"messages_dropped", drained)
			}
			return
		}
	}
}
