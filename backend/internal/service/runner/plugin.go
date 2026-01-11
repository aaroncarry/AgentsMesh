package runner

import (
	"github.com/anthropics/agentmesh/backend/internal/domain/runner"
)

// GetPluginOptions returns plugin capabilities filtered by agent type
// If agentType is empty, returns all plugins with configurable UI
func (s *Service) GetPluginOptions(r *runner.Runner, agentType string) []runner.PluginCapability {
	if r.Capabilities == nil {
		return nil
	}

	var result []runner.PluginCapability
	for _, cap := range r.Capabilities {
		// Filter by agent type if provided
		if agentType != "" {
			// Empty supported_agents means supports all agents
			if len(cap.SupportedAgents) > 0 {
				supported := false
				for _, agent := range cap.SupportedAgents {
					if agent == agentType {
						supported = true
						break
					}
				}
				if !supported {
					continue
				}
			}
		}

		// Only return plugins with configurable UI
		if cap.UI != nil && cap.UI.Configurable {
			result = append(result, cap)
		}
	}

	return result
}
