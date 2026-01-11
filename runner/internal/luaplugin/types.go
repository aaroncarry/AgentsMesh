// Package luaplugin provides a Lua-based plugin system for sandbox configuration.
package luaplugin

import "context"

// SandboxReader provides read-only access to sandbox state.
// Use this interface when plugins only need to read sandbox information.
type SandboxReader interface {
	GetPodKey() string
	GetRootPath() string
	GetWorkDir() string
	GetLaunchArgs() []string
	GetEnvVars() map[string]string
	GetMetadata() map[string]interface{}
}

// SandboxArgsWriter provides methods to modify launch arguments.
// Use this interface when plugins need to add command line arguments.
type SandboxArgsWriter interface {
	SetLaunchArgs(args []string)
	AppendLaunchArgs(args ...string) // Thread-safe atomic append
}

// SandboxEnvWriter provides methods to modify environment variables.
// Use this interface when plugins need to set environment variables.
type SandboxEnvWriter interface {
	SetEnvVar(key, value string)
}

// SandboxMetadataWriter provides methods to modify metadata.
// Use this interface when plugins need to store/retrieve metadata.
type SandboxMetadataWriter interface {
	SetMetadata(key string, value interface{})
}

// SandboxAdapter is the full interface that allows luaplugin to work with sandbox
// without creating an import cycle.
// It composes all smaller interfaces for backward compatibility.
// New code should prefer using the smaller, more focused interfaces.
type SandboxAdapter interface {
	SandboxReader
	SandboxArgsWriter
	SandboxEnvWriter
	SandboxMetadataWriter
}

// LuaPlugin represents a loaded Lua plugin.
type LuaPlugin struct {
	Name            string    `json:"name"`
	Version         string    `json:"version"`
	Description     string    `json:"description"`
	Order           int       `json:"order"`
	Critical        bool      `json:"critical"`
	SupportedAgents []string  `json:"supported_agents"`
	Executable      string    `json:"executable,omitempty"` // CLI command to check availability
	UI              *UIConfig `json:"ui,omitempty"`

	// Internal
	content     []byte // Lua script content
	isBuiltin   bool
	isAvailable bool // Whether the executable is available on this system
}

// IsBuiltin returns true if this is a builtin plugin.
func (p *LuaPlugin) IsBuiltin() bool {
	return p.isBuiltin
}

// IsAvailable returns true if the plugin's executable is available on this system.
// Plugins without an executable requirement are always considered available.
func (p *LuaPlugin) IsAvailable() bool {
	return p.isAvailable
}

// Content returns the Lua script content.
func (p *LuaPlugin) Content() []byte {
	return p.content
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

// PluginCapability represents capability info for server reporting.
type PluginCapability struct {
	Name            string    `json:"name"`
	Version         string    `json:"version"`
	Description     string    `json:"description"`
	SupportedAgents []string  `json:"supported_agents"`
	Executable      string    `json:"executable,omitempty"` // Required CLI command (if any)
	Available       bool      `json:"available"`            // Whether the executable is available on this system
	UI              *UIConfig `json:"ui,omitempty"`
}

// PluginExecutorInterface defines the contract for plugin execution.
// This interface supports dependency injection and makes testing easier.
// Implementations should handle plugin filtering, ordering, and execution.
type PluginExecutorInterface interface {
	// Execute runs all applicable plugins for the given agent type.
	// Plugins are filtered by agent type and executed in order.
	// If a critical plugin fails, execution stops and returns an error.
	Execute(ctx context.Context, plugins []*LuaPlugin, agentType string, sb SandboxAdapter, config map[string]interface{}) error

	// Teardown runs teardown for all applicable plugins in reverse order.
	// Errors are logged but do not stop other plugins from tearing down.
	Teardown(ctx context.Context, plugins []*LuaPlugin, agentType string, sb SandboxAdapter) error
}

// PluginLoaderInterface defines the contract for loading plugins.
// This interface supports dependency injection and makes testing easier.
type PluginLoaderInterface interface {
	// LoadBuiltinPlugins loads all embedded builtin plugins.
	LoadBuiltinPlugins() ([]*LuaPlugin, error)

	// LoadUserPlugins loads plugins from the specified directory.
	// loadedNames is used to skip plugins that conflict with already loaded names.
	LoadUserPlugins(dir string, loadedNames map[string]bool) ([]*LuaPlugin, error)
}

// PluginParserInterface defines the contract for parsing plugin metadata.
type PluginParserInterface interface {
	// Parse extracts metadata from a Lua plugin script.
	Parse(filename string, content []byte, isBuiltin bool) (*LuaPlugin, error)

	// Validate checks if a plugin has valid configuration.
	Validate(plugin *LuaPlugin, filename string) error
}
