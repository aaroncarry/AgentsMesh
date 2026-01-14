package runner

import (
	"github.com/anthropics/agentmesh/backend/internal/service/agent"
)

// AgentServiceAdapter adapts agent.Service to AgentTypesProvider interface
type AgentServiceAdapter struct {
	agentService *agent.Service
}

// NewAgentServiceAdapter creates a new adapter
func NewAgentServiceAdapter(agentService *agent.Service) *AgentServiceAdapter {
	return &AgentServiceAdapter{agentService: agentService}
}

// GetAgentTypesForRunner implements AgentTypesProvider interface
func (a *AgentServiceAdapter) GetAgentTypesForRunner() []AgentTypeInfo {
	// Get agent types from service
	types := a.agentService.GetAgentTypesForRunner()

	// Convert to runner.AgentTypeInfo
	result := make([]AgentTypeInfo, len(types))
	for i, t := range types {
		result[i] = AgentTypeInfo{
			Slug:          t.Slug,
			Name:          t.Name,
			Executable:    t.Executable,
			LaunchCommand: t.LaunchCommand,
		}
	}
	return result
}
