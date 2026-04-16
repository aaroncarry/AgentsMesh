package gitops

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

const (
	defaultStatusTimeout = 15 * time.Second
	defaultDiffTimeout   = 30 * time.Second
	defaultCommitTimeout = 60 * time.Second
	defaultPushTimeout   = 2 * time.Minute
	defaultDiffMaxBytes  = 256 * 1024
	defaultDiffContext   = 3
)

// Executor runs structured Git operations inside a worktree.
type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) Status(ctx context.Context, workDir string) (*runnerv1.GitStatusResult, error) {
	if err := ensureWorktree(workDir); err != nil {
		return nil, err
	}
	if err := ensureGitRepo(ctx, workDir); err != nil {
		return nil, err
	}

	statusCtx, cancel := context.WithTimeout(ctx, defaultStatusTimeout)
	defer cancel()

	branch, _ := runGit(statusCtx, workDir, nil, "rev-parse", "--abbrev-ref", "HEAD")
	headSHA, _ := runGit(statusCtx, workDir, nil, "rev-parse", "HEAD")
	porcelain, err := runGit(statusCtx, workDir, nil, "status", "--porcelain=v1")
	if err != nil {
		return nil, err
	}

	files, stats, hasStaged := parseStatusOutput(porcelain)
	return &runnerv1.GitStatusResult{
		Branch:           strings.TrimSpace(branch),
		HeadSha:          strings.TrimSpace(headSHA),
		HasChanges:       len(files) > 0,
		HasStagedChanges: hasStaged,
		Files:            files,
		Stats:            stats,
	}, nil
}

func (e *Executor) Diff(ctx context.Context, workDir string, cmd *runnerv1.GitDiffCommand) (*runnerv1.GitDiffResult, error) {
	if err := ensureWorktree(workDir); err != nil {
		return nil, err
	}
	if err := ensureGitRepo(ctx, workDir); err != nil {
		return nil, err
	}

	diffCtx, cancel := context.WithTimeout(ctx, defaultDiffTimeout)
	defer cancel()

	contextLines := int(cmd.GetContextLines())
	if contextLines <= 0 {
		contextLines = defaultDiffContext
	}
	maxBytes := int(cmd.GetMaxBytes())
	if maxBytes <= 0 {
		maxBytes = defaultDiffMaxBytes
	}

	args := []string{"diff"}
	if cmd.GetStaged() {
		args = append(args, "--cached")
	}
	args = append(args, fmt.Sprintf("--unified=%d", contextLines))
	if path := strings.TrimSpace(cmd.GetPath()); path != "" {
		if err := validateRelativePath(path); err != nil {
			return nil, err
		}
		args = append(args, "--", path)
	}

	diffOutput, err := runGit(diffCtx, workDir, nil, args...)
	if err != nil {
		return nil, err
	}
	branch, _ := runGit(diffCtx, workDir, nil, "rev-parse", "--abbrev-ref", "HEAD")

	truncated := false
	if len(diffOutput) > maxBytes {
		diffOutput = diffOutput[:maxBytes]
		truncated = true
	}

	return &runnerv1.GitDiffResult{
		Branch:    strings.TrimSpace(branch),
		Path:      strings.TrimSpace(cmd.GetPath()),
		Staged:    cmd.GetStaged(),
		Truncated: truncated,
		MaxBytes:  int32(maxBytes),
		Diff:      diffOutput,
	}, nil
}

func (e *Executor) Commit(ctx context.Context, workDir string, cmd *runnerv1.GitCommitCommand) (*runnerv1.GitCommitResult, error) {
	if err := ensureWorktree(workDir); err != nil {
		return nil, err
	}
	if err := ensureGitRepo(ctx, workDir); err != nil {
		return nil, err
	}
	if strings.TrimSpace(cmd.GetMessage()) == "" {
		return nil, newCommandError("invalid_commit_request", "commit message is required")
	}
	if cmd.GetAll() && len(cmd.GetPaths()) > 0 {
		return nil, newCommandError("invalid_commit_request", "paths and all cannot be used together")
	}
	if (cmd.GetAuthorName() == "") != (cmd.GetAuthorEmail() == "") {
		return nil, newCommandError("invalid_commit_request", "author_name and author_email must be provided together")
	}

	commitCtx, cancel := context.WithTimeout(ctx, defaultCommitTimeout)
	defer cancel()

	if cmd.GetAll() {
		if _, err := runGit(commitCtx, workDir, nil, "add", "-A"); err != nil {
			return nil, err
		}
	} else if len(cmd.GetPaths()) > 0 {
		paths := make([]string, 0, len(cmd.GetPaths()))
		for _, path := range cmd.GetPaths() {
			if err := validateRelativePath(path); err != nil {
				return nil, err
			}
			paths = append(paths, path)
		}
		args := append([]string{"add", "--"}, paths...)
		if _, err := runGit(commitCtx, workDir, nil, args...); err != nil {
			return nil, err
		}
	}

	stagedNames, err := runGit(commitCtx, workDir, nil, "diff", "--cached", "--name-only")
	if err != nil {
		return nil, err
	}
	committedFiles := parseNameList(stagedNames)
	if len(committedFiles) == 0 {
		return nil, newCommandError("no_changes", "no changes to commit")
	}

	args := make([]string, 0, 8)
	if cmd.GetAuthorName() != "" {
		args = append(args,
			"-c", "user.name="+cmd.GetAuthorName(),
			"-c", "user.email="+cmd.GetAuthorEmail(),
		)
	}
	args = append(args, "commit", "-m", cmd.GetMessage())
	if _, err := runGit(commitCtx, workDir, nil, args...); err != nil {
		if strings.Contains(err.Error(), "Author identity unknown") {
			return nil, newCommandError("author_identity_required", "git author identity is not configured")
		}
		return nil, err
	}

	branch, _ := runGit(commitCtx, workDir, nil, "rev-parse", "--abbrev-ref", "HEAD")
	headSHA, _ := runGit(commitCtx, workDir, nil, "rev-parse", "HEAD")

	return &runnerv1.GitCommitResult{
		Branch:         strings.TrimSpace(branch),
		CommitSha:      strings.TrimSpace(headSHA),
		Message:        cmd.GetMessage(),
		CommittedFiles: committedFiles,
	}, nil
}

func (e *Executor) Push(ctx context.Context, workDir string, cmd *runnerv1.GitPushCommand) (*runnerv1.GitPushResult, error) {
	if err := ensureWorktree(workDir); err != nil {
		return nil, err
	}
	if err := ensureGitRepo(ctx, workDir); err != nil {
		return nil, err
	}

	branch := strings.TrimSpace(cmd.GetBranch())
	if branch == "" {
		return nil, newCommandError("invalid_branch", "branch is required")
	}
	if err := checkBranchName(ctx, workDir, branch); err != nil {
		return nil, err
	}
	if strings.TrimSpace(cmd.GetRemoteUrl()) == "" {
		return nil, newCommandError("invalid_remote_url", "remote_url is required")
	}
	if cmd.GetAuth() == nil || strings.TrimSpace(cmd.GetAuth().GetToken()) == "" || strings.TrimSpace(cmd.GetAuth().GetUsername()) == "" {
		return nil, newCommandError("auth_required", "auth.username and auth.token are required")
	}

	pushEnv, err := buildPushEnv(cmd.GetRemoteUrl(), cmd.GetAuth().GetUsername(), cmd.GetAuth().GetToken())
	if err != nil {
		return nil, err
	}

	pushCtx, cancel := context.WithTimeout(ctx, defaultPushTimeout)
	defer cancel()

	args := []string{"push"}
	if cmd.GetSetUpstream() {
		args = append(args, "--set-upstream")
	}
	if cmd.GetForceWithLease() {
		args = append(args, "--force-with-lease")
	}
	args = append(args, cmd.GetRemoteUrl(), fmt.Sprintf("HEAD:refs/heads/%s", branch))

	if _, err := runGit(pushCtx, workDir, pushEnv, args...); err != nil {
		return nil, classifyPushError(err)
	}

	headSHA, _ := runGit(pushCtx, workDir, nil, "rev-parse", "HEAD")
	return &runnerv1.GitPushResult{
		Branch:        branch,
		RemoteUrl:     cmd.GetRemoteUrl(),
		Pushed:        true,
		UpstreamSet:   cmd.GetSetUpstream(),
		RemoteHeadSha: strings.TrimSpace(headSHA),
	}, nil
}

func ensureWorktree(workDir string) error {
	if strings.TrimSpace(workDir) == "" {
		return newCommandError("workspace_not_ready", "workspace path is not available")
	}
	info, err := os.Stat(workDir)
	if err != nil || !info.IsDir() {
		return newCommandError("workspace_not_ready", "workspace path is not available")
	}
	return nil
}

func ensureGitRepo(ctx context.Context, workDir string) error {
	repoCtx, cancel := context.WithTimeout(ctx, defaultStatusTimeout)
	defer cancel()

	out, err := runGit(repoCtx, workDir, nil, "rev-parse", "--is-inside-work-tree")
	if err != nil || strings.TrimSpace(out) != "true" {
		return newCommandError("repo_not_initialized", "workspace is not a git repository")
	}
	return nil
}

func runGit(ctx context.Context, dir string, extraEnv []string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if output == "" {
			output = err.Error()
		}
		return "", fmt.Errorf("%s", output)
	}
	return string(out), nil
}

func parseStatusOutput(output string) ([]*runnerv1.GitStatusFile, *runnerv1.GitStatusStats, bool) {
	files := []*runnerv1.GitStatusFile{}
	stats := &runnerv1.GitStatusStats{}
	hasStaged := false

	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimRight(raw, "\r")
		if len(line) < 3 {
			continue
		}
		xy := line[:2]
		path := strings.TrimSpace(line[3:])
		if idx := strings.LastIndex(path, " -> "); idx >= 0 {
			path = path[idx+4:]
		}
		status := classifyStatus(xy)
		staged := xy[0] != ' ' && xy[0] != '?'
		if staged {
			hasStaged = true
		}
		switch status {
		case "added":
			stats.Added++
		case "deleted":
			stats.Deleted++
		case "renamed":
			stats.Renamed++
		case "untracked":
			stats.Untracked++
		default:
			stats.Modified++
		}
		files = append(files, &runnerv1.GitStatusFile{
			Path:   path,
			Status: status,
			Staged: staged,
		})
	}

	return files, stats, hasStaged
}

func classifyStatus(xy string) string {
	switch {
	case xy == "??":
		return "untracked"
	case strings.ContainsRune(xy, 'R'):
		return "renamed"
	case strings.ContainsRune(xy, 'D'):
		return "deleted"
	case strings.ContainsRune(xy, 'A'):
		return "added"
	default:
		return "modified"
	}
}

func parseNameList(output string) []string {
	items := []string{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			items = append(items, line)
		}
	}
	return items
}

func validateRelativePath(path string) error {
	if strings.TrimSpace(path) == "" {
		return newCommandError("invalid_path", "path cannot be empty")
	}
	if filepath.IsAbs(path) {
		return newCommandError("invalid_path", "path must be relative")
	}
	clean := filepath.Clean(path)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return newCommandError("invalid_path", "path cannot escape workspace")
	}
	return nil
}

func checkBranchName(ctx context.Context, workDir, branch string) error {
	checkCtx, cancel := context.WithTimeout(ctx, defaultStatusTimeout)
	defer cancel()

	if _, err := runGit(checkCtx, workDir, nil, "check-ref-format", "--branch", branch); err != nil {
		return newCommandError("invalid_branch", "branch name is invalid")
	}
	return nil
}

func buildPushEnv(remoteURL, username, token string) ([]string, error) {
	parsed, err := url.Parse(remoteURL)
	if err != nil {
		return nil, newCommandError("invalid_remote_url", "remote_url is invalid")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, newCommandError("invalid_remote_url", "remote_url must use http or https")
	}
	if parsed.Host == "" {
		return nil, newCommandError("invalid_remote_url", "remote_url host is required")
	}
	if parsed.User != nil {
		return nil, newCommandError("invalid_remote_url", "remote_url must not contain embedded credentials")
	}

	baseURL := fmt.Sprintf("%s://%s/", parsed.Scheme, parsed.Host)
	authValue := base64.StdEncoding.EncodeToString([]byte(username + ":" + token))
	return []string{
		"GIT_TERMINAL_PROMPT=0",
		"GIT_CONFIG_COUNT=3",
		"GIT_CONFIG_KEY_0=credential.helper",
		"GIT_CONFIG_VALUE_0=",
		"GIT_CONFIG_KEY_1=core.askPass",
		"GIT_CONFIG_VALUE_1=",
		"GIT_CONFIG_KEY_2=http." + baseURL + ".extraHeader",
		"GIT_CONFIG_VALUE_2=Authorization: Basic " + authValue,
	}, nil
}

func classifyPushError(err error) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "Authentication failed"),
		strings.Contains(msg, "HTTP Basic: Access denied"),
		strings.Contains(msg, "could not read Username"),
		strings.Contains(msg, "access denied"):
		return newCommandError("auth_failed", "git authentication failed for this push request")
	case strings.Contains(msg, "non-fast-forward"),
		strings.Contains(msg, "[rejected]"),
		strings.Contains(msg, "fetch first"):
		return newCommandError("push_rejected_non_fast_forward", "push rejected because the remote contains newer commits")
	case strings.Contains(strings.ToLower(msg), "protected branch"):
		return newCommandError("protected_branch", "push rejected by protected branch rules")
	default:
		return newCommandError("push_failed", msg)
	}
}
