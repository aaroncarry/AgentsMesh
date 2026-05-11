package factory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/logger"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
)

const maxFactorySettingsFileSize = 2 * 1024 * 1024

type factoryParser struct{}

type factorySessionSettings struct {
	Model      string `json:"model"`
	TokenUsage struct {
		InputTokens         int64 `json:"inputTokens"`
		OutputTokens        int64 `json:"outputTokens"`
		CacheCreationTokens int64 `json:"cacheCreationTokens"`
		CacheReadTokens     int64 `json:"cacheReadTokens"`
	} `json:"tokenUsage"`
}

func (p *factoryParser) Parse(sandboxPath string, podStartedAt time.Time) (*tokenusage.TokenUsage, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		logger.Pod().Warn("Factory parser: cannot determine HOME", "error", err)
		return nil, nil
	}

	usage := tokenusage.NewTokenUsage()
	for _, sessionsDir := range factorySessionDirs(home, sandboxPath) {
		parseFactorySessionDir(sessionsDir, podStartedAt, usage)
	}

	if usage.IsEmpty() {
		return nil, nil
	}
	return usage, nil
}

func factorySessionDirs(home, sandboxPath string) []string {
	if home == "" || sandboxPath == "" {
		return nil
	}

	root := filepath.Join(home, ".factory", "sessions")
	candidates := []string{
		filepath.Join(sandboxPath, "workspace"),
		sandboxPath,
	}

	seen := make(map[string]struct{}, len(candidates)*2)
	var dirs []string
	for _, candidate := range candidates {
		for _, resolved := range factoryResolvedPathCandidates(candidate) {
			name := factorySessionDirName(resolved)
			if name == "" {
				continue
			}
			dir := filepath.Join(root, name)
			if _, ok := seen[dir]; ok {
				continue
			}
			seen[dir] = struct{}{}
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

func factoryResolvedPathCandidates(path string) []string {
	if path == "" {
		return nil
	}

	candidates := []string{filepath.Clean(path)}
	if abs, err := filepath.Abs(path); err == nil {
		candidates = append(candidates, filepath.Clean(abs))
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		candidates = append(candidates, filepath.Clean(resolved))
	}

	seen := make(map[string]struct{}, len(candidates))
	unique := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == "" || candidate == "." {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		unique = append(unique, candidate)
	}
	return unique
}

func factorySessionDirName(path string) string {
	path = filepath.Clean(path)
	if path == "" || path == "." {
		return ""
	}

	var b strings.Builder
	b.Grow(len(path))
	for _, r := range path {
		switch r {
		case '/', '\\':
			b.WriteByte('-')
		case ':':
			// Match common CLI session-dir conventions on Windows drive paths.
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func parseFactorySessionDir(dir string, podStartedAt time.Time, usage *tokenusage.TokenUsage) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".settings.json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if !tokenusage.IsModifiedAfter(path, podStartedAt) {
			continue
		}
		if err := parseFactorySettingsFile(path, usage); err != nil {
			logger.Pod().Warn("Factory parser: file parse error", "file", path, "error", err)
		}
	}
}

func parseFactorySettingsFile(path string, usage *tokenusage.TokenUsage) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() > maxFactorySettingsFileSize {
		logger.Pod().Warn("Factory parser: skipping oversized settings file", "file", path, "size", info.Size())
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var settings factorySessionSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil
	}

	model := settings.Model
	if model == "" {
		model = "factory-unknown"
	}
	u := settings.TokenUsage
	if u.InputTokens == 0 && u.OutputTokens == 0 && u.CacheCreationTokens == 0 && u.CacheReadTokens == 0 {
		return nil
	}

	usage.Add(
		model,
		u.InputTokens,
		u.OutputTokens,
		u.CacheCreationTokens,
		u.CacheReadTokens,
	)
	return nil
}
