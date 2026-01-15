package runner

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

// HandleMessage handles an incoming message from a runner
func (cm *ConnectionManager) HandleMessage(runnerID int64, msgType int, data []byte) {
	if msgType != websocket.TextMessage && msgType != websocket.BinaryMessage {
		return
	}

	var msg RunnerMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		cm.logger.Warn("failed to parse runner message",
			"runner_id", runnerID,
			"error", err)
		return
	}

	switch msg.Type {
	// Initialization flow
	case MsgTypeInitialize:
		cm.handleInitializeMessage(runnerID, msg.Data)

	case MsgTypeInitialized:
		cm.handleInitializedMessage(runnerID, msg.Data)

	// Runtime messages
	case MsgTypeHeartbeat:
		cm.handleHeartbeatMessage(runnerID, msg.Data)

	case MsgTypePodCreated:
		cm.handlePodCreatedMessage(runnerID, msg.Data)

	case MsgTypePodTerminated:
		cm.handlePodTerminatedMessage(runnerID, msg.Data)

	case MsgTypeTerminalOutput:
		cm.handleTerminalOutputMessage(runnerID, msg.Data)

	case MsgTypeAgentStatus:
		cm.handleAgentStatusMessage(runnerID, msg.Data)

	case MsgTypePtyResized:
		cm.handlePtyResizedMessage(runnerID, msg.Data)

	default:
		cm.logger.Debug("unknown message type",
			"runner_id", runnerID,
			"type", msg.Type)
	}
}

// handleHeartbeatMessage handles heartbeat message
func (cm *ConnectionManager) handleHeartbeatMessage(runnerID int64, data json.RawMessage) {
	var hbData HeartbeatData
	if err := json.Unmarshal(data, &hbData); err != nil {
		cm.logger.Error("failed to unmarshal heartbeat data",
			"runner_id", runnerID,
			"error", err,
			"data", string(data))
		return
	}

	cm.logger.Debug("received heartbeat",
		"runner_id", runnerID,
		"pods", len(hbData.Pods))
	cm.UpdateHeartbeat(runnerID)

	if cm.onHeartbeat != nil {
		cm.onHeartbeat(runnerID, &hbData)
	}
}

// handlePodCreatedMessage handles pod created message
func (cm *ConnectionManager) handlePodCreatedMessage(runnerID int64, data json.RawMessage) {
	var pcData PodCreatedData
	if err := json.Unmarshal(data, &pcData); err != nil {
		cm.logger.Error("failed to unmarshal pod_created data",
			"runner_id", runnerID,
			"error", err)
		return
	}
	if cm.onPodCreated != nil {
		cm.onPodCreated(runnerID, &pcData)
	}
}

// handlePodTerminatedMessage handles pod terminated message
func (cm *ConnectionManager) handlePodTerminatedMessage(runnerID int64, data json.RawMessage) {
	var ptData PodTerminatedData
	if err := json.Unmarshal(data, &ptData); err != nil {
		cm.logger.Error("failed to unmarshal pod_terminated data",
			"runner_id", runnerID,
			"error", err)
		return
	}
	if cm.onPodTerminated != nil {
		cm.onPodTerminated(runnerID, &ptData)
	}
}

// handleTerminalOutputMessage handles terminal output message
func (cm *ConnectionManager) handleTerminalOutputMessage(runnerID int64, data json.RawMessage) {
	var toData TerminalOutputData
	if err := json.Unmarshal(data, &toData); err != nil {
		cm.logger.Error("failed to unmarshal terminal_output data",
			"runner_id", runnerID,
			"error", err)
		return
	}
	if cm.onTerminalOutput != nil {
		cm.onTerminalOutput(runnerID, &toData)
	}
}

// handleAgentStatusMessage handles agent status message
func (cm *ConnectionManager) handleAgentStatusMessage(runnerID int64, data json.RawMessage) {
	var asData AgentStatusData
	if err := json.Unmarshal(data, &asData); err != nil {
		cm.logger.Error("failed to unmarshal agent_status data",
			"runner_id", runnerID,
			"error", err)
		return
	}
	if cm.onAgentStatus != nil {
		cm.onAgentStatus(runnerID, &asData)
	}
}

// handlePtyResizedMessage handles PTY resized message
func (cm *ConnectionManager) handlePtyResizedMessage(runnerID int64, data json.RawMessage) {
	var prData PtyResizedData
	if err := json.Unmarshal(data, &prData); err != nil {
		cm.logger.Error("failed to unmarshal pty_resized data",
			"runner_id", runnerID,
			"error", err)
		return
	}
	if cm.onPtyResized != nil {
		cm.onPtyResized(runnerID, &prData)
	}
}

// ========== Initialization Handlers ==========

// failInitialization handles initialization failure - logs, notifies callback, and cleans up connection
func (cm *ConnectionManager) failInitialization(runnerID int64, reason string) {
	cm.logger.Error("initialization failed",
		"runner_id", runnerID,
		"reason", reason)

	// Notify callback before removing connection
	if cm.onInitFailed != nil {
		cm.onInitFailed(runnerID, reason)
	}

	// Clean up the connection
	cm.RemoveConnection(runnerID)
}

// handleInitializeMessage handles the initialize message from runner (phase 1)
func (cm *ConnectionManager) handleInitializeMessage(runnerID int64, data json.RawMessage) {
	var params InitializeParams
	if err := json.Unmarshal(data, &params); err != nil {
		cm.failInitialization(runnerID, "invalid initialize params: "+err.Error())
		return
	}

	cm.logger.Info("received initialize request",
		"runner_id", runnerID,
		"protocol_version", params.ProtocolVersion,
		"runner_version", params.RunnerInfo.Version,
		"mcp_port", params.RunnerInfo.MCPPort,
		"os", params.RunnerInfo.OS,
		"arch", params.RunnerInfo.Arch)

	// Protocol version check
	if params.ProtocolVersion < MinSupportedProtocolVersion {
		cm.logger.Warn("runner protocol version too old",
			"runner_id", runnerID,
			"runner_version", params.ProtocolVersion,
			"min_supported", MinSupportedProtocolVersion)
		// Still send response but log warning - could reject in strict mode
	}
	if params.ProtocolVersion > CurrentProtocolVersion {
		cm.logger.Info("runner using newer protocol version",
			"runner_id", runnerID,
			"runner_version", params.ProtocolVersion,
			"server_version", CurrentProtocolVersion)
	}

	// Get agent types from provider
	var agentTypes []AgentTypeInfo
	if cm.agentTypesProvider != nil {
		agentTypes = cm.agentTypesProvider.GetAgentTypesForRunner()
	}

	if len(agentTypes) == 0 {
		cm.logger.Warn("no agent types available for runner",
			"runner_id", runnerID)
	}

	// Build initialize_result response
	result := InitializeResult{
		ProtocolVersion: CurrentProtocolVersion,
		ServerInfo: ServerInfo{
			Version: cm.serverVersion,
		},
		AgentTypes: agentTypes,
		Features:   SupportedFeatures(),
	}

	// Send initialize_result (phase 2)
	resultData, err := json.Marshal(result)
	if err != nil {
		cm.failInitialization(runnerID, "failed to marshal initialize_result: "+err.Error())
		return
	}

	msg := &RunnerMessage{
		Type:      MsgTypeInitializeResult,
		Data:      resultData,
		Timestamp: time.Now().UnixMilli(),
	}

	if err := cm.SendMessage(context.Background(), runnerID, msg); err != nil {
		cm.failInitialization(runnerID, "failed to send initialize_result: "+err.Error())
		return
	}

	cm.logger.Info("sent initialize_result",
		"runner_id", runnerID,
		"agent_types", len(agentTypes),
		"features", result.Features)
}

// handleInitializedMessage handles the initialized message from runner (phase 3)
func (cm *ConnectionManager) handleInitializedMessage(runnerID int64, data json.RawMessage) {
	var params InitializedParams
	if err := json.Unmarshal(data, &params); err != nil {
		cm.logger.Error("failed to unmarshal initialized params",
			"runner_id", runnerID,
			"error", err)
		return
	}

	cm.logger.Info("runner initialization completed",
		"runner_id", runnerID,
		"available_agents", params.AvailableAgents)

	// Update connection state (thread-safe)
	conn := cm.GetConnection(runnerID)
	if conn != nil {
		conn.SetInitialized(true, params.AvailableAgents)
	}

	// Notify callback
	if cm.onInitialized != nil {
		cm.onInitialized(runnerID, params.AvailableAgents)
	}
}
