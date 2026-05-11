package agent

import (
	"embed"
	"path"
	"strings"

	extensionservice "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

//go:embed builtin_skills/*.md
var builtinSkillsFS embed.FS

const resourcePlaceholderWorkDir = "{{.sandbox.work_dir}}"
const resourcePlaceholderRootPath = "{{.sandbox.root_path}}"

func appendAgentSkillFiles(files []*runnerv1.FileToCreate, agentSlug string, skillSlugs []string) []*runnerv1.FileToCreate {
	skillRoot := skillRootForAgent(agentSlug)
	if skillRoot == "" {
		return files
	}

	seen := make(map[string]struct{}, len(skillSlugs))
	for _, slug := range skillSlugs {
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}

		content, ok := builtinSkillContent(slug)
		if !ok {
			continue
		}
		skillDir := path.Join(skillRoot, slug)
		files = append(files,
			&runnerv1.FileToCreate{Path: skillDir, IsDirectory: true},
			&runnerv1.FileToCreate{Path: path.Join(skillDir, "SKILL.md"), Content: content, Mode: 0644},
		)
	}
	return files
}

func buildSkillResources(agentSlug string, skills []*extensionservice.ResolvedSkill) []*runnerv1.ResourceToDownload {
	configRoot := skillResourceConfigRootForAgent(agentSlug)
	if configRoot == "" || len(skills) == 0 {
		return nil
	}

	resources := make([]*runnerv1.ResourceToDownload, 0, len(skills))
	for _, skill := range skills {
		if skill == nil || skill.ContentSha == "" {
			continue
		}

		targetDir := skill.TargetDir
		if targetDir == "" {
			targetDir = path.Join("skills", skill.Slug)
		}
		targetDir = path.Clean(targetDir)
		if targetDir == "." || path.IsAbs(targetDir) || targetDir == ".." || strings.HasPrefix(targetDir, "../") {
			continue
		}

		resources = append(resources, &runnerv1.ResourceToDownload{
			Sha:          skill.ContentSha,
			DownloadUrl:  skill.DownloadURL,
			TargetPath:   path.Join(configRoot, targetDir),
			ResourceType: "skill_package",
			SizeBytes:    skill.PackageSize,
		})
	}
	return resources
}

func builtinSkillContent(slug string) (string, bool) {
	data, err := builtinSkillsFS.ReadFile(path.Join("builtin_skills", slug+".md"))
	if err != nil {
		return "", false
	}
	return string(data), true
}

func skillRootForAgent(agentSlug string) string {
	if isFactoryCLI(agentSlug) {
		return path.Join(PlaceholderWorkDir, ".factory", "skills")
	}
	return ""
}

func skillResourceConfigRootForAgent(agentSlug string) string {
	switch canonicalSkillAgentSlug(agentSlug) {
	case "factory-cli":
		return path.Join(resourcePlaceholderWorkDir, ".factory")
	case "codex-cli":
		return path.Join(resourcePlaceholderRootPath, "codex-home")
	case "claude-code":
		return path.Join(resourcePlaceholderWorkDir, ".claude")
	default:
		return resourcePlaceholderRootPath
	}
}

func isFactoryCLI(agentSlug string) bool {
	return canonicalSkillAgentSlug(agentSlug) == "factory-cli"
}

func canonicalSkillAgentSlug(agentSlug string) string {
	switch agentSlug {
	case "factory-droid":
		return "factory-cli"
	case "codex":
		return "codex-cli"
	case "claude":
		return "claude-code"
	default:
		return agentSlug
	}
}
