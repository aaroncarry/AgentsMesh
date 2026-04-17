package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/gitops"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

type gitResultSender interface {
	SendGitCommandResult(result *runnerv1.GitCommandResult) error
}

// OnGitCommand handles structured git commands from server and reports the result back via gRPC.
func (h *RunnerMessageHandler) OnGitCommand(cmd *runnerv1.GitCommand) error {
	log := logger.Pod()

	sender, ok := h.conn.(gitResultSender)
	if !ok {
		return fmt.Errorf("connection does not support SendGitCommandResult")
	}

	result := &runnerv1.GitCommandResult{
		RequestId: cmd.GetRequestId(),
		PodKey:    cmd.GetPodKey(),
	}

	workDir, err := h.resolveGitWorkDir(cmd)
	if err != nil {
		applyGitError(result, err)
		return sender.SendGitCommandResult(result)
	}

	executor := gitops.NewExecutor()
	ctx := context.Background()

	switch action := cmd.GetAction().(type) {
	case *runnerv1.GitCommand_Status:
		status, execErr := executor.Status(ctx, workDir)
		if execErr != nil {
			applyGitError(result, execErr)
			break
		}
		result.Ok = true
		result.Result = &runnerv1.GitCommandResult_Status{Status: status}

	case *runnerv1.GitCommand_Diff:
		diff, execErr := executor.Diff(ctx, workDir, action.Diff)
		if execErr != nil {
			applyGitError(result, execErr)
			break
		}
		result.Ok = true
		result.Result = &runnerv1.GitCommandResult_Diff{Diff: diff}

	case *runnerv1.GitCommand_Commit:
		commit, execErr := executor.Commit(ctx, workDir, action.Commit)
		if execErr != nil {
			applyGitError(result, execErr)
			break
		}
		result.Ok = true
		result.Result = &runnerv1.GitCommandResult_Commit{Commit: commit}

	case *runnerv1.GitCommand_Push:
		push, execErr := executor.Push(ctx, workDir, action.Push)
		if execErr != nil {
			applyGitError(result, execErr)
			break
		}
		result.Ok = true
		result.Result = &runnerv1.GitCommandResult_Push{Push: push}

	default:
		result.Ok = false
		result.Code = "invalid_git_command"
		result.Message = "git command action is required"
	}

	log.Info("Git command completed",
		"request_id", result.GetRequestId(),
		"pod_key", result.GetPodKey(),
		"ok", result.GetOk(),
		"code", result.GetCode(),
	)
	return sender.SendGitCommandResult(result)
}

func (h *RunnerMessageHandler) resolveGitWorkDir(cmd *runnerv1.GitCommand) (string, error) {
	if pod, ok := h.podStore.Get(cmd.GetPodKey()); ok && strings.TrimSpace(pod.WorkDir) != "" {
		return pod.WorkDir, nil
	}

	sandboxPath := strings.TrimSpace(cmd.GetSandboxPath())
	if sandboxPath == "" {
		return "", &gitops.CommandError{Code: "workspace_not_ready", Message: "workspace path is not available"}
	}

	workspacePath := filepath.Join(sandboxPath, "workspace")
	if info, err := os.Stat(workspacePath); err == nil && info.IsDir() {
		return workspacePath, nil
	}
	if info, err := os.Stat(sandboxPath); err == nil && info.IsDir() {
		return sandboxPath, nil
	}
	return "", &gitops.CommandError{Code: "workspace_not_ready", Message: "workspace path is not available"}
}

func applyGitError(result *runnerv1.GitCommandResult, err error) {
	if cmdErr, ok := err.(*gitops.CommandError); ok {
		result.Ok = false
		result.Code = cmdErr.Code
		result.Message = cmdErr.Message
		return
	}
	result.Ok = false
	result.Code = "internal_error"
	result.Message = err.Error()
}
