package luaplugin

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/anthropics/agentmesh/runner/internal/luaplugin/builtin"
	lua "github.com/yuin/gopher-lua"
)

// PluginExecutor handles executing Lua plugins.
// It implements PluginExecutorInterface.
// Shared modules are loaded once per executor instance instead of using global state.
type PluginExecutor struct {
	sharedModules map[string][]byte
	modulesOnce   sync.Once
	modulesErr    error
}

// Compile-time check that PluginExecutor implements PluginExecutorInterface
var _ PluginExecutorInterface = (*PluginExecutor)(nil)

// NewPluginExecutor creates a new PluginExecutor.
func NewPluginExecutor() *PluginExecutor {
	return &PluginExecutor{
		sharedModules: make(map[string][]byte),
	}
}

// initSharedModules loads shared modules once per executor.
// This replaces the global init() function for better testability.
func (e *PluginExecutor) initSharedModules() error {
	e.modulesOnce.Do(func() {
		// Load mcp_utils.lua as shared module
		content, err := builtin.BuiltinPlugins.ReadFile("mcp_utils.lua")
		if err != nil {
			e.modulesErr = fmt.Errorf("failed to load mcp_utils.lua: %w", err)
			return
		}
		e.sharedModules["mcp_utils"] = content
	})
	return e.modulesErr
}

// Execute runs all applicable plugins for the given agent type.
func (e *PluginExecutor) Execute(ctx context.Context, plugins []*LuaPlugin, agentType string, sb SandboxAdapter, config map[string]interface{}) error {
	// Filter plugins for this agent type
	applicable := e.filterPlugins(plugins, agentType)

	// Sort by order
	sort.Slice(applicable, func(i, j int) bool {
		return applicable[i].Order < applicable[j].Order
	})

	log.Printf("[luaplugin] Executing %d plugins for agent type: %s", len(applicable), agentType)

	// Execute each plugin
	for _, plugin := range applicable {
		if err := e.executePlugin(ctx, plugin, sb, config); err != nil {
			if plugin.Critical {
				return fmt.Errorf("critical plugin %s failed: %w", plugin.Name, err)
			}
			log.Printf("[luaplugin] Warning: non-critical plugin %s failed: %v", plugin.Name, err)
		}
	}

	return nil
}

// Teardown runs teardown for all applicable plugins (in reverse order).
func (e *PluginExecutor) Teardown(ctx context.Context, plugins []*LuaPlugin, agentType string, sb SandboxAdapter) error {
	applicable := e.filterPlugins(plugins, agentType)

	// Sort in reverse order
	sort.Slice(applicable, func(i, j int) bool {
		return applicable[i].Order > applicable[j].Order
	})

	for _, plugin := range applicable {
		if err := e.teardownPlugin(ctx, plugin, sb); err != nil {
			log.Printf("[luaplugin] Warning: teardown for %s failed: %v", plugin.Name, err)
		}
	}

	return nil
}

// filterPlugins returns plugins applicable to the given agent type.
func (e *PluginExecutor) filterPlugins(plugins []*LuaPlugin, agentType string) []*LuaPlugin {
	var result []*LuaPlugin
	for _, p := range plugins {
		// Empty supported_agents means supports all agents
		if len(p.SupportedAgents) == 0 {
			result = append(result, p)
			continue
		}

		// Check if this agent type is supported
		for _, supported := range p.SupportedAgents {
			if supported == agentType {
				result = append(result, p)
				break
			}
		}
	}
	return result
}

// filterConfigForPlugin extracts plugin-specific config by stripping the plugin name prefix.
// Frontend sends namespaced keys like "claude-code.skip_permissions" to avoid conflicts
// between plugins. This function extracts only the keys for the specified plugin and
// removes the prefix so Lua plugins can access config as ctx.config.skip_permissions.
// Non-namespaced keys (legacy support) are preserved as-is.
func filterConfigForPlugin(config map[string]interface{}, pluginName string) map[string]interface{} {
	if config == nil {
		return nil
	}

	prefix := pluginName + "."
	result := make(map[string]interface{})

	for k, v := range config {
		if strings.HasPrefix(k, prefix) {
			// Strip prefix: "claude-code.skip_permissions" → "skip_permissions"
			newKey := strings.TrimPrefix(k, prefix)
			result[newKey] = v
		} else if !strings.Contains(k, ".") {
			// Keep non-namespaced keys (legacy support): "ticket_identifier", "working_dir"
			result[k] = v
		}
		// Ignore other plugins' namespaced config
	}

	return result
}

// loadSharedModules loads shared Lua modules into the state.
func (e *PluginExecutor) loadSharedModules(L *lua.LState) error {
	// Initialize shared modules on first use
	if err := e.initSharedModules(); err != nil {
		return err
	}

	for name, content := range e.sharedModules {
		if err := L.DoString(string(content)); err != nil {
			return fmt.Errorf("failed to load shared module %s: %w", name, err)
		}
	}
	return nil
}

// executePlugin runs a single Lua plugin.
// The context is used to control execution timeout - if the context is cancelled,
// the Lua execution will be interrupted.
func (e *PluginExecutor) executePlugin(ctx context.Context, plugin *LuaPlugin, sb SandboxAdapter, config map[string]interface{}) error {
	L := lua.NewState()
	defer L.Close()

	// Set context for timeout/cancellation support.
	// gopher-lua will check context cancellation periodically during execution.
	L.SetContext(ctx)

	// Load shared modules first (mcp_utils, etc.)
	if err := e.loadSharedModules(L); err != nil {
		return err
	}

	// Load and execute plugin (APIs are provided via ctx table, not global registration)
	if err := L.DoString(string(plugin.content)); err != nil {
		// Check if the error is due to context cancellation
		if ctx.Err() != nil {
			return fmt.Errorf("lua execution cancelled: %w", ctx.Err())
		}
		return fmt.Errorf("lua execution error: %w", err)
	}

	// Call setup function
	setupFn := L.GetGlobal("setup")
	if setupFn == lua.LNil {
		log.Printf("[luaplugin] Plugin %s has no setup function, skipping", plugin.Name)
		return nil
	}

	// Filter config for this specific plugin (strip namespace prefix)
	pluginConfig := filterConfigForPlugin(config, plugin.Name)

	// Debug: log original and filtered config
	log.Printf("[luaplugin] Plugin %s - original config: %+v", plugin.Name, config)
	log.Printf("[luaplugin] Plugin %s - filtered config: %+v", plugin.Name, pluginConfig)

	// Create context table with filtered config
	ctxTable := createContextTable(L, sb, pluginConfig)

	// Call setup(ctx)
	if err := L.CallByParam(lua.P{
		Fn:      setupFn,
		NRet:    0,
		Protect: true,
	}, ctxTable); err != nil {
		// Check if the error is due to context cancellation
		if ctx.Err() != nil {
			return fmt.Errorf("setup() cancelled: %w", ctx.Err())
		}
		return fmt.Errorf("setup() failed: %w", err)
	}

	log.Printf("[luaplugin] Plugin %s executed successfully", plugin.Name)
	return nil
}

// teardownPlugin runs teardown for a single plugin.
// The context is used to control execution timeout.
func (e *PluginExecutor) teardownPlugin(ctx context.Context, plugin *LuaPlugin, sb SandboxAdapter) error {
	L := lua.NewState()
	defer L.Close()

	// Set context for timeout/cancellation support
	L.SetContext(ctx)

	// Load shared modules first
	if err := e.loadSharedModules(L); err != nil {
		return err
	}

	// Load plugin (APIs are provided via ctx table, not global registration)
	if err := L.DoString(string(plugin.content)); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("lua execution cancelled: %w", ctx.Err())
		}
		return err
	}

	teardownFn := L.GetGlobal("teardown")
	if teardownFn == lua.LNil {
		return nil
	}

	ctxTable := createContextTable(L, sb, nil)

	if err := L.CallByParam(lua.P{
		Fn:      teardownFn,
		NRet:    0,
		Protect: true,
	}, ctxTable); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("teardown() cancelled: %w", ctx.Err())
		}
		return err
	}

	return nil
}
