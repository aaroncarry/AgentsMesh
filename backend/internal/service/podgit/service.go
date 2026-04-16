package podgit

import (
	"context"
	"errors"
	"strings"
	"time"

	agentpoddomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	runnersvc "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

type podLookup interface {
	GetPod(ctx context.Context, podKey string) (*agentpoddomain.Pod, error)
}

type gitExecutor interface {
	Execute(ctx context.Context, runnerID int64, cmd *runnerv1.GitCommand) (*runnerv1.GitCommandResult, error)
}

// Service orchestrates Pod-scoped Git operations through RunnerCommandSender.
type Service struct {
	pods     podLookup
	executor gitExecutor
}

func NewService(pods podLookup, executor gitExecutor) *Service {
	return &Service{
		pods:     pods,
		executor: executor,
	}
}

func (s *Service) Status(ctx context.Context, orgID int64, podKey string) (*StatusResponse, error) {
	pod, err := s.getPodForGit(ctx, orgID, podKey)
	if err != nil {
		return nil, err
	}

	result, err := s.execute(ctx, pod, &runnerv1.GitCommand{
		PodKey:      pod.PodKey,
		SandboxPath: strings.TrimSpace(derefString(pod.SandboxPath)),
		Action:      &runnerv1.GitCommand_Status{Status: &runnerv1.GitStatusCommand{}},
	}, 30*time.Second)
	if err != nil {
		return nil, err
	}

	status := result.GetStatus()
	return &StatusResponse{
		Ok:               true,
		PodKey:           pod.PodKey,
		Branch:           status.GetBranch(),
		HeadSHA:          status.GetHeadSha(),
		HasChanges:       status.GetHasChanges(),
		HasStagedChanges: status.GetHasStagedChanges(),
		Files:            status.GetFiles(),
		Stats:            status.GetStats(),
	}, nil
}

func (s *Service) Diff(ctx context.Context, orgID int64, podKey string, req DiffRequest) (*DiffResponse, error) {
	pod, err := s.getPodForGit(ctx, orgID, podKey)
	if err != nil {
		return nil, err
	}

	result, err := s.execute(ctx, pod, &runnerv1.GitCommand{
		PodKey:      pod.PodKey,
		SandboxPath: strings.TrimSpace(derefString(pod.SandboxPath)),
		Action: &runnerv1.GitCommand_Diff{Diff: &runnerv1.GitDiffCommand{
			Path:         req.Path,
			Staged:       req.Staged,
			ContextLines: req.Context,
			MaxBytes:     req.MaxBytes,
		}},
	}, 60*time.Second)
	if err != nil {
		return nil, err
	}

	diff := result.GetDiff()
	return &DiffResponse{
		Ok:        true,
		PodKey:    pod.PodKey,
		Branch:    diff.GetBranch(),
		Path:      diff.GetPath(),
		Staged:    diff.GetStaged(),
		Truncated: diff.GetTruncated(),
		MaxBytes:  diff.GetMaxBytes(),
		Diff:      diff.GetDiff(),
	}, nil
}

func (s *Service) Commit(ctx context.Context, orgID int64, podKey string, req CommitRequest) (*CommitResponse, error) {
	pod, err := s.getPodForGit(ctx, orgID, podKey)
	if err != nil {
		return nil, err
	}

	cmd := &runnerv1.GitCommitCommand{
		Message: req.Message,
		Paths:   req.Paths,
		All:     req.All,
	}
	if req.Author != nil {
		cmd.AuthorName = req.Author.Name
		cmd.AuthorEmail = req.Author.Email
	}

	result, err := s.execute(ctx, pod, &runnerv1.GitCommand{
		PodKey:      pod.PodKey,
		SandboxPath: strings.TrimSpace(derefString(pod.SandboxPath)),
		Action:      &runnerv1.GitCommand_Commit{Commit: cmd},
	}, 90*time.Second)
	if err != nil {
		return nil, err
	}

	commit := result.GetCommit()
	return &CommitResponse{
		Ok:             true,
		PodKey:         pod.PodKey,
		Branch:         commit.GetBranch(),
		CommitSHA:      commit.GetCommitSha(),
		Message:        commit.GetMessage(),
		CommittedFiles: commit.GetCommittedFiles(),
	}, nil
}

func (s *Service) Push(ctx context.Context, orgID int64, podKey string, req PushRequest) (*PushResponse, error) {
	pod, err := s.getPodForGit(ctx, orgID, podKey)
	if err != nil {
		return nil, err
	}

	result, err := s.execute(ctx, pod, &runnerv1.GitCommand{
		PodKey:      pod.PodKey,
		SandboxPath: strings.TrimSpace(derefString(pod.SandboxPath)),
		Action: &runnerv1.GitCommand_Push{Push: &runnerv1.GitPushCommand{
			Branch:         req.Branch,
			RemoteUrl:      req.RemoteURL,
			SetUpstream:    req.SetUpstream,
			ForceWithLease: req.ForceWithLease,
			Auth: &runnerv1.GitAuth{
				Username: req.Auth.Username,
				Token:    req.Auth.Token,
			},
		}},
	}, 3*time.Minute)
	if err != nil {
		return nil, err
	}

	push := result.GetPush()
	return &PushResponse{
		Ok:            true,
		PodKey:        pod.PodKey,
		Branch:        push.GetBranch(),
		RemoteURL:     push.GetRemoteUrl(),
		Pushed:        push.GetPushed(),
		UpstreamSet:   push.GetUpstreamSet(),
		RemoteHeadSHA: push.GetRemoteHeadSha(),
	}, nil
}

func (s *Service) getPodForGit(ctx context.Context, orgID int64, podKey string) (*agentpoddomain.Pod, error) {
	if s.pods == nil {
		return nil, internalError("pod service is not configured")
	}

	pod, err := s.pods.GetPod(ctx, podKey)
	if err != nil {
		return nil, notFound("Pod not found")
	}
	if pod.OrganizationID != orgID {
		return nil, forbidden("Access denied")
	}
	if pod.RunnerID == 0 {
		return nil, serviceUnavailable("RUNNER_NOT_CONNECTED", "Runner is not connected")
	}
	if pod.SandboxPath == nil || strings.TrimSpace(*pod.SandboxPath) == "" {
		return nil, badRequest("WORKSPACE_NOT_READY", "Pod workspace is not ready")
	}
	return pod, nil
}

func (s *Service) execute(ctx context.Context, pod *agentpoddomain.Pod, cmd *runnerv1.GitCommand, timeout time.Duration) (*runnerv1.GitCommandResult, error) {
	if s.executor == nil {
		return nil, serviceUnavailable("SERVICE_UNAVAILABLE", "Pod Git executor is not configured")
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := s.executor.Execute(execCtx, pod.RunnerID, cmd)
	if err != nil {
		return nil, mapExecuteError(err)
	}
	if result == nil {
		return nil, internalError("empty git command result")
	}
	if !result.GetOk() {
		return nil, mapResultError(result)
	}
	return result, nil
}

func mapExecuteError(err error) error {
	switch {
	case errors.Is(err, runnersvc.ErrCommandSenderNotSet):
		return serviceUnavailable("SERVICE_UNAVAILABLE", "Runner command sender is not configured")
	case errors.Is(err, context.DeadlineExceeded):
		return gatewayTimeout("COMMAND_TIMEOUT", "Git command timed out")
	case errors.Is(err, context.Canceled):
		return gatewayTimeout("COMMAND_CANCELED", "Git command was canceled")
	case grpcstatus.Code(err) == codes.NotFound:
		return serviceUnavailable("RUNNER_NOT_CONNECTED", "Runner is not connected")
	default:
		return internalError("Failed to execute git command")
	}
}

func mapResultError(result *runnerv1.GitCommandResult) error {
	code := strings.ToUpper(strings.TrimSpace(result.GetCode()))
	if code == "" {
		code = "INTERNAL_ERROR"
	}
	message := strings.TrimSpace(result.GetMessage())
	if message == "" {
		message = "Git command failed"
	}
	switch code {
	case "COMMAND_TIMEOUT":
		return gatewayTimeout(code, message)
	default:
		return badRequest(code, message)
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
