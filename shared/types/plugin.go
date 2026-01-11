// Package types provides shared type definitions used across runner and backend.
package types

// PluginCapability represents capability info for server reporting.
// This type is shared between runner and backend to ensure consistency.
type PluginCapability struct {
	Name            string    `json:"name"`
	Version         string    `json:"version"`
	Description     string    `json:"description"`
	SupportedAgents []string  `json:"supported_agents"`
	UI              *UIConfig `json:"ui,omitempty"`
}

// UIConfig represents the UI configuration for a plugin.
type UIConfig struct {
	Configurable bool      `json:"configurable"`
	Fields       []UIField `json:"fields"`
}

// UIField represents a single UI field configuration.
type UIField struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // boolean, string, select, number, secret
	Label       string      `json:"label"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description,omitempty"`
	Placeholder string      `json:"placeholder,omitempty"`
	Options     []UIOption  `json:"options,omitempty"`
	Min         *float64    `json:"min,omitempty"`
	Max         *float64    `json:"max,omitempty"`
	Required    bool        `json:"required,omitempty"`
}

// UIOption represents an option for select fields.
type UIOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}
