package sandbox

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// Manager manages sandbox lifecycle for pods.
type Manager struct {
	sandboxesDir string             // Directory where sandboxes are created
	reposDir     string             // Directory for repository cache
	mcpPort      int                // MCP HTTP Server port
	plugins      []Plugin           // Registered plugins
	sandboxes    map[string]*Sandbox // Active sandboxes by pod key
	mu           sync.RWMutex
}

// NewManager creates a new SandboxManager.
func NewManager(workspace string, mcpPort int) *Manager {
	if mcpPort == 0 {
		mcpPort = 19000 // Default MCP port
	}

	return &Manager{
		sandboxesDir: filepath.Join(workspace, "sandboxes"),
		reposDir:     filepath.Join(workspace, "repos"),
		mcpPort:      mcpPort,
		plugins:      make([]Plugin, 0),
		sandboxes:    make(map[string]*Sandbox),
	}
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

	// Execute plugin Setup in order
	for _, p := range sortedPlugins {
		log.Printf("[sandbox] Running plugin %s for pod %s", p.Name(), podKey)
		if err := p.Setup(ctx, sb, config); err != nil {
			// Rollback: teardown plugins that were successfully set up
			m.teardownPlugins(sb)
			// Clean up directory
			os.RemoveAll(sandboxPath)
			return nil, fmt.Errorf("plugin %s setup failed: %w", p.Name(), err)
		}
		sb.AddPlugin(p)
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
