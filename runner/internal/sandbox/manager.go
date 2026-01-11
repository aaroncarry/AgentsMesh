package sandbox

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/anthropics/agentmesh/runner/internal/luaplugin"
)

// Manager manages sandbox lifecycle for pods.
type Manager struct {
	sandboxesDir     string                   // Directory where sandboxes are created
	reposDir         string                   // Directory for repository cache
	mcpPort          int                      // MCP HTTP Server port
	plugins          []Plugin                 // Registered Go plugins
	luaPluginManager *luaplugin.PluginManager // Lua plugin manager
	sandboxes        map[string]*Sandbox      // Active sandboxes by pod key
	mu               sync.RWMutex
}

// ManagerConfig holds configuration options for SandboxManager.
type ManagerConfig struct {
	Workspace      string // Workspace root for sandboxes and repos cache
	MCPPort        int    // MCP HTTP Server port
	UserPluginsDir string // User custom plugins directory
}

// NewManager creates a new SandboxManager.
func NewManager(workspace string, mcpPort int) *Manager {
	return NewManagerWithConfig(ManagerConfig{
		Workspace: workspace,
		MCPPort:   mcpPort,
	})
}

// NewManagerWithConfig creates a new SandboxManager with full configuration.
func NewManagerWithConfig(cfg ManagerConfig) *Manager {
	if cfg.MCPPort == 0 {
		cfg.MCPPort = 19000 // Default MCP port
	}

	// Initialize Lua plugin manager with user plugins directory
	luaPluginMgr := luaplugin.NewPluginManager(cfg.UserPluginsDir)
	if err := luaPluginMgr.LoadPlugins(); err != nil {
		log.Printf("[sandbox] Warning: failed to load Lua plugins: %v", err)
	} else {
		plugins := luaPluginMgr.GetPlugins()
		builtinCount := 0
		userCount := 0
		for _, p := range plugins {
			if p.IsBuiltin() {
				builtinCount++
			} else {
				userCount++
			}
		}
		log.Printf("[sandbox] Loaded %d Lua plugins (%d builtin, %d user)", len(plugins), builtinCount, userCount)
	}

	return &Manager{
		sandboxesDir:     filepath.Join(cfg.Workspace, "sandboxes"),
		reposDir:         filepath.Join(cfg.Workspace, "repos"),
		mcpPort:          cfg.MCPPort,
		plugins:          make([]Plugin, 0),
		luaPluginManager: luaPluginMgr,
		sandboxes:        make(map[string]*Sandbox),
	}
}

// GetLuaPluginCapabilities returns capabilities of loaded Lua plugins.
func (m *Manager) GetLuaPluginCapabilities() []luaplugin.PluginCapability {
	if m.luaPluginManager == nil {
		return nil
	}
	return m.luaPluginManager.GetCapabilities()
}

// GetLuaPluginManager returns the Lua plugin manager.
func (m *Manager) GetLuaPluginManager() *luaplugin.PluginManager {
	return m.luaPluginManager
}

// RegisterPlugin adds a plugin to the manager.
// Plugins are sorted by Order() when Create is called.
func (m *Manager) RegisterPlugin(p Plugin) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.plugins = append(m.plugins, p)
}

// GetReposDir returns the repository cache directory.
func (m *Manager) GetReposDir() string {
	return m.reposDir
}

// GetMCPPort returns the MCP HTTP Server port.
func (m *Manager) GetMCPPort() int {
	return m.mcpPort
}

// Create creates a new sandbox for the given pod.
func (m *Manager) Create(ctx context.Context, podKey string, config map[string]interface{}) (*Sandbox, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if sandbox already exists
	if sb, exists := m.sandboxes[podKey]; exists {
		return sb, nil
	}

	// Create sandbox directory
	sandboxPath := filepath.Join(m.sandboxesDir, podKey)
	if err := os.MkdirAll(sandboxPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sandbox directory: %w", err)
	}

	// Create sandbox instance
	sb := NewSandbox(podKey, sandboxPath)

	// Sort plugins by order
	sortedPlugins := make([]Plugin, len(m.plugins))
	copy(sortedPlugins, m.plugins)
	sort.Slice(sortedPlugins, func(i, j int) bool {
		return sortedPlugins[i].Order() < sortedPlugins[j].Order()
	})

	// Execute Go plugin Setup in order
	for _, p := range sortedPlugins {
		log.Printf("[sandbox] Running Go plugin %s for pod %s", p.Name(), podKey)
		if err := p.Setup(ctx, sb, config); err != nil {
			// Rollback: teardown plugins that were successfully set up
			m.teardownPlugins(sb)
			// Clean up directory
			os.RemoveAll(sandboxPath)
			return nil, fmt.Errorf("plugin %s setup failed: %w", p.Name(), err)
		}
		sb.AddPlugin(p)
	}

	// Execute Lua plugins
	if m.luaPluginManager != nil {
		// Get agent type from config
		agentType := GetStringConfig(config, "agent_type")
		if agentType == "" {
			agentType = "claude-code" // Default agent type
		}

		// Merge mcp_port into config for Lua plugins
		luaConfig := make(map[string]interface{})
		for k, v := range config {
			luaConfig[k] = v
		}
		luaConfig["mcp_port"] = m.mcpPort

		log.Printf("[sandbox] Running Lua plugins for agent type: %s", agentType)
		if err := m.luaPluginManager.Execute(ctx, agentType, sb, luaConfig); err != nil {
			m.teardownPlugins(sb)
			os.RemoveAll(sandboxPath)
			return nil, fmt.Errorf("lua plugin execution failed: %w", err)
		}
	}

	// Save sandbox metadata
	if err := sb.Save(); err != nil {
		log.Printf("[sandbox] Warning: failed to save sandbox metadata: %v", err)
	}

	// Store sandbox
	m.sandboxes[podKey] = sb

	log.Printf("[sandbox] Created sandbox for pod %s at %s", podKey, sandboxPath)
	return sb, nil
}

// Get retrieves an existing sandbox by pod key.
func (m *Manager) Get(podKey string) (*Sandbox, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sb, exists := m.sandboxes[podKey]
	return sb, exists
}

// Cleanup removes a sandbox and runs plugin Teardown.
func (m *Manager) Cleanup(podKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sb, exists := m.sandboxes[podKey]
	if !exists {
		return nil // Already cleaned up
	}

	// Run plugin teardown in reverse order
	m.teardownPlugins(sb)

	// Remove sandbox directory
	if err := os.RemoveAll(sb.RootPath); err != nil {
		log.Printf("[sandbox] Warning: failed to remove sandbox directory %s: %v", sb.RootPath, err)
	}

	// Remove from map
	delete(m.sandboxes, podKey)

	log.Printf("[sandbox] Cleaned up sandbox for pod %s", podKey)
	return nil
}

// teardownPlugins runs Teardown on all plugins in reverse order.
func (m *Manager) teardownPlugins(sb *Sandbox) {
	plugins := sb.GetPlugins()
	for i := len(plugins) - 1; i >= 0; i-- {
		p := plugins[i]
		log.Printf("[sandbox] Tearing down plugin %s for pod %s", p.Name(), sb.PodKey)
		if err := p.Teardown(sb); err != nil {
			log.Printf("[sandbox] Warning: plugin %s teardown failed: %v", p.Name(), err)
		}
	}
}

// LoadExisting loads an existing sandbox from disk.
func (m *Manager) LoadExisting(podKey string) (*Sandbox, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already loaded
	if sb, exists := m.sandboxes[podKey]; exists {
		return sb, nil
	}

	sandboxPath := filepath.Join(m.sandboxesDir, podKey)
	sb := &Sandbox{
		RootPath: sandboxPath,
		plugins:  make([]Plugin, 0),
	}

	if err := sb.Load(); err != nil {
		return nil, fmt.Errorf("failed to load sandbox: %w", err)
	}

	m.sandboxes[podKey] = sb
	return sb, nil
}

// List returns all active sandbox pod keys.
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.sandboxes))
	for k := range m.sandboxes {
		keys = append(keys, k)
	}
	return keys
}
