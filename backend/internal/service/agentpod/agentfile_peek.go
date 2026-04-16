package agentpod

import (
	"github.com/anthropics/agentsmesh/agentfile/extract"
	"github.com/anthropics/agentsmesh/agentfile/parser"
)

// peekRepoSlug extracts the REPO slug from an AgentFile source without full merge/resolve.
// Returns empty string if no REPO is declared or if parsing fails.
func peekRepoSlug(agentfileSrc string) string {
	if agentfileSrc == "" {
		return ""
	}
	prog, errs := parser.Parse(agentfileSrc)
	if len(errs) > 0 || prog == nil {
		return ""
	}
	spec := extract.Extract(prog)
	if spec.Repo != nil && spec.Repo.URL != "" {
		return spec.Repo.URL
	}
	return ""
}
