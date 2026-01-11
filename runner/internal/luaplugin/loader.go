package luaplugin

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anthropics/agentmesh/runner/internal/luaplugin/builtin"
)

// PluginLoader handles loading plugins from various sources.
type PluginLoader struct {
	parser *PluginParser
}

// NewPluginLoader creates a new PluginLoader.
func NewPluginLoader() *PluginLoader {
	return &PluginLoader{
		parser: NewPluginParser(),
	}
}

// LoadBuiltinPlugins loads all builtin plugins from embedded filesystem.
func (l *PluginLoader) LoadBuiltinPlugins() ([]*LuaPlugin, error) {
	var plugins []*LuaPlugin

	entries, err := builtin.BuiltinPlugins.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read builtin plugins: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".lua") {
			content, err := builtin.BuiltinPlugins.ReadFile(entry.Name())
			if err != nil {
				log.Printf("[luaplugin] Warning: failed to read builtin plugin %s: %v", entry.Name(), err)
				continue
			}

			plugin, err := l.loadFromContent(entry.Name(), content, true)
			if err != nil {
				log.Printf("[luaplugin] Warning: failed to load builtin plugin %s: %v", entry.Name(), err)
				continue
			}

			plugins = append(plugins, plugin)
			log.Printf("[luaplugin] Loaded builtin plugin: %s (order=%d)", plugin.Name, plugin.Order)
		}
	}

	return plugins, nil
}

// LoadUserPlugins loads plugins from a user-specified directory.
// loadedNames is used to skip plugins that conflict with already loaded ones.
func (l *PluginLoader) LoadUserPlugins(dir string, loadedNames map[string]bool) ([]*LuaPlugin, error) {
	var plugins []*LuaPlugin

	// Check if directory exists
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		log.Printf("[luaplugin] User plugins directory does not exist: %s", dir)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat user plugins directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("user plugins path is not a directory: %s", dir)
	}

	// Find all .lua files
	pattern := filepath.Join(dir, "*.lua")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob user plugins: %w", err)
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Printf("[luaplugin] Warning: failed to read user plugin %s: %v", file, err)
			continue
		}

		plugin, err := l.loadFromContent(filepath.Base(file), content, false)
		if err != nil {
			log.Printf("[luaplugin] Warning: failed to load user plugin %s: %v", file, err)
			continue
		}

		// Check for conflicts
		if loadedNames[plugin.Name] {
			log.Printf("[luaplugin] Warning: user plugin '%s' conflicts with builtin plugin, skipping", plugin.Name)
			continue
		}

		plugins = append(plugins, plugin)
		log.Printf("[luaplugin] Loaded user plugin: %s from %s (order=%d)", plugin.Name, file, plugin.Order)
	}

	return plugins, nil
}

// loadFromContent parses and validates a plugin from content.
func (l *PluginLoader) loadFromContent(filename string, content []byte, isBuiltin bool) (*LuaPlugin, error) {
	plugin, err := l.parser.Parse(filename, content, isBuiltin)
	if err != nil {
		return nil, err
	}

	if err := l.parser.Validate(plugin, filename); err != nil {
		return nil, err
	}

	// Check if the required executable is available
	l.checkExecutable(plugin)

	return plugin, nil
}

// checkExecutable checks if the plugin's required executable is available.
// Updates the plugin's isAvailable field based on the result.
func (l *PluginLoader) checkExecutable(plugin *LuaPlugin) {
	// Plugins without an executable requirement are always available
	if plugin.Executable == "" {
		plugin.isAvailable = true
		return
	}

	// Check if the executable exists in PATH
	_, err := exec.LookPath(plugin.Executable)
	plugin.isAvailable = (err == nil)

	if !plugin.isAvailable {
		log.Printf("[luaplugin] Plugin '%s' unavailable: executable '%s' not found in PATH",
			plugin.Name, plugin.Executable)
	} else {
		log.Printf("[luaplugin] Plugin '%s' available: executable '%s' found",
			plugin.Name, plugin.Executable)
	}
}
