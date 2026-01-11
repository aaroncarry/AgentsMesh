// Package luaplugin provides a Lua-based plugin system for sandbox configuration.
package luaplugin

import (
	"context"
	"log"
)

// PluginManager manages Lua plugins for sandbox configuration.
// It coordinates between loader and executor components.
// Dependencies can be injected via functional options for better testability.
type PluginManager struct {
	userPluginsDir string
	plugins        []*LuaPlugin
	loader         PluginLoaderInterface
	executor       PluginExecutorInterface
}

// Option is a functional option for configuring PluginManager.
type Option func(*PluginManager)

// WithLoader sets a custom loader implementation.
// Useful for testing with mock loaders.
func WithLoader(loader PluginLoaderInterface) Option {
	return func(m *PluginManager) {
		m.loader = loader
	}
}

// WithExecutor sets a custom executor implementation.
// Useful for testing with mock executors.
func WithExecutor(executor PluginExecutorInterface) Option {
	return func(m *PluginManager) {
		m.executor = executor
	}
}

// NewPluginManager creates a new PluginManager.
// Use functional options to inject dependencies for testing.
func NewPluginManager(userPluginsDir string, opts ...Option) *PluginManager {
	m := &PluginManager{
		userPluginsDir: userPluginsDir,
		plugins:        make([]*LuaPlugin, 0),
	}

	// Apply options first
	for _, opt := range opts {
		opt(m)
	}

	// Set defaults for any unset dependencies
	if m.loader == nil {
		m.loader = NewPluginLoader()
	}
	if m.executor == nil {
		m.executor = NewPluginExecutor()
	}

	return m
}

// LoadPlugins loads all builtin and user plugins.
func (m *PluginManager) LoadPlugins() error {
	// Track loaded plugin names to avoid duplicates
	loadedNames := make(map[string]bool)

	// 1. Load builtin plugins
	builtinPlugins, err := m.loader.LoadBuiltinPlugins()
	if err != nil {
		return err
	}

	for _, plugin := range builtinPlugins {
		m.plugins = append(m.plugins, plugin)
		loadedNames[plugin.Name] = true
	}

	// 2. Load user plugins
	if m.userPluginsDir != "" {
		userPlugins, err := m.loader.LoadUserPlugins(m.userPluginsDir, loadedNames)
		if err != nil {
			// Don't fail on user plugin loading errors, just log a warning
			log.Printf("[luaplugin] Warning: failed to load user plugins: %v", err)
		}

		for _, plugin := range userPlugins {
			m.plugins = append(m.plugins, plugin)
			loadedNames[plugin.Name] = true
		}
	}

	return nil
}

// GetPlugins returns all loaded plugins.
func (m *PluginManager) GetPlugins() []*LuaPlugin {
	return m.plugins
}

// GetCapabilities returns plugin capabilities for reporting to server.
// All plugins are included with their availability status.
func (m *PluginManager) GetCapabilities() []PluginCapability {
	var caps []PluginCapability
	for _, p := range m.plugins {
		caps = append(caps, PluginCapability{
			Name:            p.Name,
			Version:         p.Version,
			Description:     p.Description,
			SupportedAgents: p.SupportedAgents,
			Executable:      p.Executable,
			Available:       p.IsAvailable(),
			UI:              p.UI,
		})
	}
	return caps
}

// Execute runs all applicable plugins for the given agent type.
func (m *PluginManager) Execute(ctx context.Context, agentType string, sb SandboxAdapter, config map[string]interface{}) error {
	return m.executor.Execute(ctx, m.plugins, agentType, sb, config)
}

// Teardown runs teardown for all plugins (in reverse order).
func (m *PluginManager) Teardown(ctx context.Context, agentType string, sb SandboxAdapter) error {
	return m.executor.Teardown(ctx, m.plugins, agentType, sb)
}
