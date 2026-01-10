package plugins

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anthropics/agentmesh/runner/internal/sandbox"
)

func TestSkillsPluginName(t *testing.T) {
	p := NewSkillsPlugin()
	if p.Name() != "skills" {
		t.Errorf("Name() = %s, want skills", p.Name())
	}
}

func TestSkillsPluginOrder(t *testing.T) {
	p := NewSkillsPlugin()
	if p.Order() != 60 {
		t.Errorf("Order() = %d, want 60", p.Order())
	}
}

func TestSkillsPluginSetup(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "skills-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create sandbox and set WorkDir (normally set by TempDirPlugin or WorktreePlugin)
	sb := sandbox.NewSandbox("test-pod", tmpDir)
	sb.WorkDir = tmpDir

	// Run plugin setup
	p := NewSkillsPlugin()
	if err := p.Setup(context.Background(), sb, nil); err != nil {
		t.Fatalf("Setup() failed: %v", err)
	}

	// Verify am-delegate skill was created
	delegateSkillPath := filepath.Join(tmpDir, ".claude", "skills", "am-delegate", "SKILL.md")
	if _, err := os.Stat(delegateSkillPath); os.IsNotExist(err) {
		t.Error("am-delegate/SKILL.md was not created")
	}

	// Verify am-delegate .gitignore was created
	delegateGitignorePath := filepath.Join(tmpDir, ".claude", "skills", "am-delegate", ".gitignore")
	if _, err := os.Stat(delegateGitignorePath); os.IsNotExist(err) {
		t.Error("am-delegate/.gitignore was not created")
	}

	// Verify am-channel skill was created
	channelSkillPath := filepath.Join(tmpDir, ".claude", "skills", "am-channel", "SKILL.md")
	if _, err := os.Stat(channelSkillPath); os.IsNotExist(err) {
		t.Error("am-channel/SKILL.md was not created")
	}

	// Verify am-channel .gitignore was created
	channelGitignorePath := filepath.Join(tmpDir, ".claude", "skills", "am-channel", ".gitignore")
	if _, err := os.Stat(channelGitignorePath); os.IsNotExist(err) {
		t.Error("am-channel/.gitignore was not created")
	}
}

func TestSkillsPluginDelegateContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skills-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sb := sandbox.NewSandbox("test-pod", tmpDir)
	sb.WorkDir = tmpDir
	p := NewSkillsPlugin()
	if err := p.Setup(context.Background(), sb, nil); err != nil {
		t.Fatalf("Setup() failed: %v", err)
	}

	// Read and verify am-delegate content
	delegateContent, err := os.ReadFile(filepath.Join(tmpDir, ".claude", "skills", "am-delegate", "SKILL.md"))
	if err != nil {
		t.Fatalf("Failed to read am-delegate/SKILL.md: %v", err)
	}

	content := string(delegateContent)

	// Check frontmatter
	if !strings.Contains(content, "name: am-delegate") {
		t.Error("am-delegate SKILL.md missing 'name: am-delegate' in frontmatter")
	}
	if !strings.Contains(content, "user-invocable: false") {
		t.Error("am-delegate SKILL.md missing 'user-invocable: false' in frontmatter")
	}

	// Check key tools are mentioned
	if !strings.Contains(content, "create_pod") {
		t.Error("am-delegate SKILL.md should mention create_pod")
	}
	if !strings.Contains(content, "list_available_pods") {
		t.Error("am-delegate SKILL.md should mention list_available_pods")
	}
	if !strings.Contains(content, "send_terminal_text") {
		t.Error("am-delegate SKILL.md should mention send_terminal_text")
	}
}

func TestSkillsPluginChannelContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skills-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sb := sandbox.NewSandbox("test-pod", tmpDir)
	sb.WorkDir = tmpDir
	p := NewSkillsPlugin()
	if err := p.Setup(context.Background(), sb, nil); err != nil {
		t.Fatalf("Setup() failed: %v", err)
	}

	// Read and verify am-channel content
	channelContent, err := os.ReadFile(filepath.Join(tmpDir, ".claude", "skills", "am-channel", "SKILL.md"))
	if err != nil {
		t.Fatalf("Failed to read am-channel/SKILL.md: %v", err)
	}

	content := string(channelContent)

	// Check frontmatter
	if !strings.Contains(content, "name: am-channel") {
		t.Error("am-channel SKILL.md missing 'name: am-channel' in frontmatter")
	}
	if !strings.Contains(content, "user-invocable: false") {
		t.Error("am-channel SKILL.md missing 'user-invocable: false' in frontmatter")
	}

	// Check key tools are mentioned
	if !strings.Contains(content, "search_channels") {
		t.Error("am-channel SKILL.md should mention search_channels")
	}
	if !strings.Contains(content, "get_channel_messages") {
		t.Error("am-channel SKILL.md should mention get_channel_messages")
	}
	if !strings.Contains(content, "send_channel_message") {
		t.Error("am-channel SKILL.md should mention send_channel_message")
	}
	if !strings.Contains(content, "update_ticket") {
		t.Error("am-channel SKILL.md should mention update_ticket")
	}
}

func TestSkillsPluginGitignoreContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skills-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sb := sandbox.NewSandbox("test-pod", tmpDir)
	sb.WorkDir = tmpDir
	p := NewSkillsPlugin()
	if err := p.Setup(context.Background(), sb, nil); err != nil {
		t.Fatalf("Setup() failed: %v", err)
	}

	// Check .gitignore content
	gitignoreContent, err := os.ReadFile(filepath.Join(tmpDir, ".claude", "skills", "am-delegate", ".gitignore"))
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}

	if string(gitignoreContent) != "*\n" {
		t.Errorf(".gitignore content = %q, want %q", string(gitignoreContent), "*\n")
	}
}

func TestSkillsPluginTeardown(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skills-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sb := sandbox.NewSandbox("test-pod", tmpDir)
	p := NewSkillsPlugin()

	// Teardown should not error
	if err := p.Teardown(sb); err != nil {
		t.Errorf("Teardown() failed: %v", err)
	}
}

func TestSkillsPluginDoesNotOverwriteExistingSkills(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skills-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create existing user skill
	userSkillDir := filepath.Join(tmpDir, ".claude", "skills", "user-skill")
	if err := os.MkdirAll(userSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create user skill dir: %v", err)
	}
	userSkillContent := "# User's custom skill\nThis should not be affected."
	if err := os.WriteFile(filepath.Join(userSkillDir, "SKILL.md"), []byte(userSkillContent), 0644); err != nil {
		t.Fatalf("Failed to write user skill: %v", err)
	}

	// Run plugin setup
	sb := sandbox.NewSandbox("test-pod", tmpDir)
	sb.WorkDir = tmpDir
	p := NewSkillsPlugin()
	if err := p.Setup(context.Background(), sb, nil); err != nil {
		t.Fatalf("Setup() failed: %v", err)
	}

	// Verify user skill is still intact
	content, err := os.ReadFile(filepath.Join(userSkillDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("Failed to read user skill: %v", err)
	}
	if string(content) != userSkillContent {
		t.Error("User's custom skill was modified")
	}

	// Verify no .gitignore was added to user's skill directory
	userGitignorePath := filepath.Join(userSkillDir, ".gitignore")
	if _, err := os.Stat(userGitignorePath); !os.IsNotExist(err) {
		t.Error(".gitignore should not be added to user's skill directory")
	}
}
