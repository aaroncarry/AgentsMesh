package runner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

// prepareAgentHome copies user's agent config directory to a per-pod isolated
// directory when CODEX_HOME is set in EnvVars. This enables per-pod MCP config
// isolation without modifying the user's original ~/.codex/ directory.
//
// After copying, it merges platform MCP servers from FilesToCreate into the
// existing config.toml using TOML-aware merging (only mcp_servers section).
func (b *PodBuilder) prepareAgentHome(sandboxRoot, workDir string) error {
	if b.cmd == nil || b.cmd.EnvVars == nil {
		return nil
	}

	codexHome, ok := b.cmd.EnvVars["CODEX_HOME"]
	if !ok || codexHome == "" {
		return nil
	}

	// Resolve template variables in CODEX_HOME path
	codexHome = b.resolvePath(codexHome, sandboxRoot, workDir)

	log := logger.Pod()
	log.Info("Preparing CODEX_HOME", "pod_key", b.cmd.PodKey, "codex_home", codexHome)

	// Copy user's ~/.codex/ to per-pod codex-home (if it exists)
	home := userHomeDir()
	if home != "" {
		userCodexDir := filepath.Join(home, ".codex")
		if dirExists(userCodexDir) {
			if err := copyDirSelective(userCodexDir, codexHome); err != nil {
				log.Warn("Failed to copy user codex dir, creating empty",
					"source", userCodexDir, "dest", codexHome, "error", err)
				// Clean up partial copy before creating empty directory
				_ = os.RemoveAll(codexHome)
				if mkErr := os.MkdirAll(codexHome, 0755); mkErr != nil {
					return fmt.Errorf("failed to create codex-home: %w", mkErr)
				}
			}
		} else {
			if err := os.MkdirAll(codexHome, 0755); err != nil {
				return fmt.Errorf("failed to create codex-home: %w", err)
			}
		}
	} else {
		if err := os.MkdirAll(codexHome, 0755); err != nil {
			return fmt.Errorf("failed to create codex-home: %w", err)
		}
	}

	// Find config.toml entry in FilesToCreate and merge it with existing config
	configTomlPath := filepath.Join(codexHome, "config.toml")
	mergeIdx := -1
	for i, f := range b.cmd.FilesToCreate {
		resolvedPath := b.resolvePath(f.Path, sandboxRoot, workDir)
		if resolvedPath == configTomlPath && !f.IsDirectory {
			mergeIdx = i
			break
		}
	}
	if mergeIdx >= 0 {
		f := b.cmd.FilesToCreate[mergeIdx]
		if err := mergeTomlMcpServers(configTomlPath, f.Content); err != nil {
			log.Warn("Failed to merge TOML MCP config, writing fresh",
				"path", configTomlPath, "error", err)
			// Fall through to let createFiles write it
		} else {
			// Remove from FilesToCreate to prevent createFiles from overwriting
			b.cmd.FilesToCreate = append(b.cmd.FilesToCreate[:mergeIdx], b.cmd.FilesToCreate[mergeIdx+1:]...)
			log.Info("Merged MCP config into existing config.toml", "path", configTomlPath)
		}
	}

	return nil
}

// mergeTomlMcpServers merges platform MCP server config into an existing config.toml.
// Only the mcp_servers section is merged; all other user settings are preserved.
func mergeTomlMcpServers(configPath, platformContent string) error {
	// Parse platform MCP content
	var platformConfig map[string]interface{}
	if err := toml.Unmarshal([]byte(platformContent), &platformConfig); err != nil {
		return fmt.Errorf("failed to parse platform TOML: %w", err)
	}

	platformServers, _ := platformConfig["mcp_servers"].(map[string]interface{})
	if len(platformServers) == 0 {
		return nil // Nothing to merge
	}

	// Read existing config (if any)
	var existingConfig map[string]interface{}
	existingData, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No existing config, write platform content directly
			return os.WriteFile(configPath, []byte(platformContent), 0644)
		}
		return fmt.Errorf("failed to read existing config: %w", err)
	}

	if err := toml.Unmarshal(existingData, &existingConfig); err != nil {
		return fmt.Errorf("failed to parse existing config: %w", err)
	}

	// Merge mcp_servers: platform entries override existing ones with same key
	existingServers, _ := existingConfig["mcp_servers"].(map[string]interface{})
	if existingServers == nil {
		existingServers = make(map[string]interface{})
	}
	for k, v := range platformServers {
		existingServers[k] = v
	}
	existingConfig["mcp_servers"] = existingServers

	// Write back
	merged, err := toml.Marshal(existingConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal merged config: %w", err)
	}

	return os.WriteFile(configPath, merged, 0644)
}

// userHomeDir returns the user's home directory, falling back gracefully.
func userHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// dirExists checks if a directory exists.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// copyDirSelective copies a directory recursively, skipping large/transient
// subdirectories (sessions, cache) that are not needed per-pod.
// Symlinks are preserved as symlinks rather than dereferenced.
// Special files (sockets, pipes, devices) are silently skipped.
// Individual file errors are logged and skipped rather than aborting the entire copy.
func copyDirSelective(src, dst string) error {
	log := logger.Pod()

	skipDirs := map[string]bool{
		"sessions": true, // Session logs can be large
		"cache":    true, // Cache is transient
	}

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Permission denied on a subdirectory — skip it, don't abort
			log.Debug("Skipping inaccessible path during copy", "path", path, "error", err)
			return nil
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dst, relPath)

		// Handle symlinks before directory checks, since WalkDir does not
		// follow symlinks and d.IsDir() returns false for symlink-to-dir.
		if d.Type()&fs.ModeSymlink != 0 {
			if symlinkErr := copySymlink(path, destPath); symlinkErr != nil {
				log.Debug("Skipping uncopiable symlink", "path", path, "error", symlinkErr)
			}
			return nil
		}

		// Skip transient directories
		if d.IsDir() {
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) > 0 && skipDirs[parts[0]] {
				return filepath.SkipDir
			}
			if mkErr := os.MkdirAll(destPath, 0755); mkErr != nil {
				log.Debug("Skipping uncreatable directory", "path", destPath, "error", mkErr)
			}
			return nil
		}

		// Skip special files: sockets, pipes, devices.
		// Only copy regular files (d.Type() == 0 means regular file).
		if !d.Type().IsRegular() {
			return nil
		}

		// Skip files larger than 10 MiB to avoid OOM on large binaries/databases
		info, err := d.Info()
		if err != nil {
			return nil
		}
		const maxFileSize = 10 << 20 // 10 MiB
		if info.Size() > maxFileSize {
			log.Debug("Skipping oversized file during copy", "path", path, "size", info.Size())
			return nil
		}

		// Copy regular file — skip on error to preserve partial results
		data, err := os.ReadFile(path)
		if err != nil {
			log.Debug("Skipping unreadable file during copy", "path", path, "error", err)
			return nil
		}

		if writeErr := os.WriteFile(destPath, data, info.Mode()); writeErr != nil {
			log.Debug("Skipping unwritable file during copy", "dest", destPath, "error", writeErr)
		}
		return nil
	})
}

// copySymlink recreates a symlink at dst pointing to the same target as src.
// Dangling symlinks are silently skipped.
func copySymlink(src, dst string) error {
	target, err := os.Readlink(src)
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	return os.Symlink(target, dst)
}
