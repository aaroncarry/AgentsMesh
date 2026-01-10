package plugins

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/anthropics/agentmesh/runner/internal/sandbox"
)

// SkillsPlugin injects collaboration skills into pod workspace.
// These skills enable agents to delegate tasks and communicate via channels.
type SkillsPlugin struct{}

// NewSkillsPlugin creates a new SkillsPlugin.
func NewSkillsPlugin() *SkillsPlugin {
	return &SkillsPlugin{}
}

func (p *SkillsPlugin) Name() string {
	return "skills"
}

func (p *SkillsPlugin) Order() int {
	return 60 // After MCPPlugin (50)
}

func (p *SkillsPlugin) Setup(ctx context.Context, sb *sandbox.Sandbox, config map[string]interface{}) error {
	skillsDir := filepath.Join(sb.WorkDir, ".claude", "skills")

	// Create am-delegate skill
	if err := p.createSkill(skillsDir, "am-delegate", DelegateSkillContent); err != nil {
		return err
	}

	// Create am-channel skill
	if err := p.createSkill(skillsDir, "am-channel", ChannelSkillContent); err != nil {
		return err
	}

	log.Printf("[skills] Injected collaboration skills at %s", skillsDir)
	return nil
}

// createSkill creates a skill directory with SKILL.md and .gitignore
func (p *SkillsPlugin) createSkill(skillsDir, name, content string) error {
	skillDir := filepath.Join(skillsDir, name)

	// Create skill directory
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return err
	}

	// Write SKILL.md
	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
		return err
	}

	// Write .gitignore to exclude this directory from git
	gitignoreFile := filepath.Join(skillDir, ".gitignore")
	if err := os.WriteFile(gitignoreFile, []byte("*\n"), 0644); err != nil {
		return err
	}

	return nil
}

func (p *SkillsPlugin) Teardown(sb *sandbox.Sandbox) error {
	// No cleanup needed - files will be removed with sandbox
	return nil
}
