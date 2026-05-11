package factory

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
)

var epoch = time.Time{}

func setHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", dir)
	}
}

func TestFactoryParser_ParseWorkspaceSettings(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)

	sandbox := t.TempDir()
	workspace := filepath.Join(sandbox, "workspace")
	require.NoError(t, os.MkdirAll(workspace, 0o755))

	sessionDir := filepath.Join(home, ".factory", "sessions", factorySessionDirName(mustEvalSymlinks(t, workspace)))
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	settings := `{
  "model": "claude-sonnet-4-5-20250929",
  "tokenUsage": {
    "inputTokens": 103,
    "outputTokens": 796,
    "cacheCreationTokens": 49870,
    "cacheReadTokens": 116436,
    "thinkingTokens": 12
  }
}`
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "session.settings.json"), []byte(settings), 0o644))

	parser := &factoryParser{}
	usage, err := parser.Parse(sandbox, epoch)
	require.NoError(t, err)
	require.NotNil(t, usage)

	m := usage.Models["claude-sonnet-4-5-20250929"]
	require.NotNil(t, m)
	assert.Equal(t, int64(103), m.InputTokens)
	assert.Equal(t, int64(796), m.OutputTokens)
	assert.Equal(t, int64(49870), m.CacheCreationTokens)
	assert.Equal(t, int64(116436), m.CacheReadTokens)
}

func TestFactoryParser_ParseDirectWorkDir(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)

	workDir := t.TempDir()
	sessionDir := filepath.Join(home, ".factory", "sessions", factorySessionDirName(mustEvalSymlinks(t, workDir)))
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionDir, "session.settings.json"),
		[]byte(`{"model":"glm-4.7","tokenUsage":{"inputTokens":10,"outputTokens":5}}`),
		0o644,
	))

	parser := &factoryParser{}
	usage, err := parser.Parse(workDir, epoch)
	require.NoError(t, err)
	require.NotNil(t, usage)
	assert.Equal(t, int64(10), usage.Models["glm-4.7"].InputTokens)
	assert.Equal(t, int64(5), usage.Models["glm-4.7"].OutputTokens)
}

func TestFactoryParser_ParseSkipsOldSettings(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)

	workspace := filepath.Join(t.TempDir(), "workspace")
	require.NoError(t, os.MkdirAll(workspace, 0o755))
	sessionDir := filepath.Join(home, ".factory", "sessions", factorySessionDirName(mustEvalSymlinks(t, workspace)))
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	settingsPath := filepath.Join(sessionDir, "old.settings.json")
	require.NoError(t, os.WriteFile(
		settingsPath,
		[]byte(`{"model":"m","tokenUsage":{"inputTokens":1,"outputTokens":1}}`),
		0o644,
	))
	oldTime := time.Now().Add(-1 * time.Hour)
	require.NoError(t, os.Chtimes(settingsPath, oldTime, oldTime))

	parser := &factoryParser{}
	usage, err := parser.Parse(filepath.Dir(workspace), time.Now().Add(-1*time.Minute))
	require.NoError(t, err)
	assert.Nil(t, usage)
}

func TestFactoryParser_ParseMalformedSettings(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "bad.settings.json")
	require.NoError(t, os.WriteFile(file, []byte("not json"), 0o644))

	usage := tokenusage.NewTokenUsage()
	err := parseFactorySettingsFile(file, usage)
	require.NoError(t, err)
	assert.True(t, usage.IsEmpty())
}

func TestFactorySessionDirName(t *testing.T) {
	assert.Equal(t, "-private-tmp-agent-workspace", factorySessionDirName("/private/tmp/agent/workspace"))
	assert.Equal(t, "C-Users-aaron-workspace", factorySessionDirName(`C:\Users\aaron\workspace`))
}

func mustEvalSymlinks(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	require.NoError(t, err)
	return resolved
}
