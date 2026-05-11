package agentpod

import (
	"fmt"
	"sort"
	"strings"

	"github.com/anthropics/agentsmesh/agentfile/extract"
	"github.com/anthropics/agentsmesh/agentfile/merge"
	"github.com/anthropics/agentsmesh/agentfile/parser"
	"github.com/anthropics/agentsmesh/agentfile/resolve"
	"github.com/anthropics/agentsmesh/agentfile/serialize"
	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
)

// agentfileExtractResult holds values extracted from a merged AgentFile (base + user layer).
// It contains both overrides for DB write and the serialized merged source for Runner,
// eliminating the need for downstream re-parsing.
type agentfileExtractResult struct {
	// Overrides for DB write
	Mode              string // MODE pty/acp
	CredentialProfile string // CREDENTIAL "profile-name"
	Branch            string // BRANCH "branch-name"
	RepoSlug          string // REPO "slug" (e.g., "dev-org/demo-api")
	PermissionMode    string // CONFIG permission_mode = "bypassPermissions"
	Prompt            string // PROMPT "prompt content"
	// Merged AgentFile source (for Runner, avoids re-parsing in ConfigBuilder).
	// CONFIG declarations contain final resolved values (post-resolve).
	MergedAgentfileSource string
	// ConfigValues captures resolved user-facing CONFIG values for pod persistence
	// and resume inheritance. System-injected values are excluded.
	ConfigValues agentDomain.ConfigValues
}

// extractFromAgentfileLayer parses the agent base AgentFile and user layer,
// merges them, resolves CONFIG values, serializes the result, and extracts declarations.
// Single-pass: parse + merge + resolve + serialize + extract — all in one place.
func extractFromAgentfileLayer(
	baseAgentfileSrc, userLayerSrc string,
	userPrefs, systemOverrides map[string]interface{},
) (*agentfileExtractResult, error) {
	baseProg, baseErrs := parser.Parse(baseAgentfileSrc)
	if len(baseErrs) > 0 {
		return nil, fmt.Errorf("base agentfile parse error: %v", baseErrs[0])
	}

	userProg, userErrs := parser.Parse(userLayerSrc)
	if len(userErrs) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAgentfileLayer, userErrs[0])
	}

	// Track which CONFIG fields were explicitly set in the user's Layer.
	layerConfigNames := resolve.ExtractConfigNames(userProg)

	merge.Merge(baseProg, userProg)

	// Inject final config values: system > layer > userPrefs > base defaults.
	resolve.ResolveConfigValues(baseProg, layerConfigNames, userPrefs, systemOverrides)

	mergedSource := serialize.Serialize(baseProg)
	spec := extract.Extract(baseProg)

	result := &agentfileExtractResult{
		Mode:                  spec.Mode,
		CredentialProfile:     spec.CredentialProfile,
		Prompt:                spec.Prompt,
		MergedAgentfileSource: mergedSource,
		ConfigValues:          make(agentDomain.ConfigValues),
	}

	if spec.Repo != nil {
		result.RepoSlug = spec.Repo.URL
		result.Branch = spec.Repo.Branch
	}

	for _, cfg := range spec.Config {
		if cfg.Default != nil && !isSystemConfigName(cfg.Name) {
			result.ConfigValues[cfg.Name] = cfg.Default
		}
		if cfg.Name == "permission_mode" {
			if s, ok := cfg.Default.(string); ok {
				result.PermissionMode = s
			}
		}
	}

	return result, nil
}

func configValuesToAgentfileLayer(configValues map[string]interface{}) string {
	if len(configValues) == 0 {
		return ""
	}

	keys := make([]string, 0, len(configValues))
	for k, v := range configValues {
		if k == "" || v == nil || isSystemConfigName(k) {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var lines []string
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("CONFIG %s = %s", k, serialize.FormatValue(configValues[k])))
	}
	return strings.Join(lines, "\n")
}

func isSystemConfigName(name string) bool {
	switch name {
	case "session_id", "resume_enabled", "resume_session":
		return true
	default:
		return false
	}
}
