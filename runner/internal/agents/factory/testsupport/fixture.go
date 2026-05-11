// Package testsupport provides testing-only fixture helpers for the Factory
// agent token usage parser.
package testsupport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureSettings = `{
  "model": "claude-sonnet-4-5-20250929",
  "tokenUsage": {
    "inputTokens": 103,
    "outputTokens": 796,
    "cacheCreationTokens": 49870,
    "cacheReadTokens": 116436
  }
}`

// BuildFixtureSandbox plants a Factory session settings file under a
// temporary HOME using the same path-derived sessions layout that Droid uses.
func BuildFixtureSandbox(t *testing.T) string {
	t.Helper()
	sandbox := t.TempDir()
	workspace := filepath.Join(sandbox, "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("factory fixture: mkdir workspace: %v", err)
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	resolved, err := filepath.EvalSymlinks(workspace)
	if err != nil {
		t.Fatalf("factory fixture: eval symlinks: %v", err)
	}
	sessionDir := filepath.Join(home, ".factory", "sessions", sessionDirName(resolved))
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("factory fixture: mkdir session dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionDir, "session.settings.json"), []byte(fixtureSettings), 0o644); err != nil {
		t.Fatalf("factory fixture: write settings: %v", err)
	}
	return sandbox
}

func sessionDirName(path string) string {
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
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
