package runner

import (
	"context"
	"sync"
	"time"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/google/uuid"
)

// GitCommandTimeout is the default timeout for git commands executed on runners.
const GitCommandTimeout = 5 * time.Minute

// GitCommandSender is a narrow interface for sending git commands to runners.
// Separated from RunnerCommandSender to avoid modifying the shared interface.
type GitCommandSender interface {
	SendGitCommand(ctx context.Context, runnerID int64, cmd *runnerv1.GitCommand) error
}

type pendingGitCommand struct {
	resultCh chan *runnerv1.GitCommandResult
	timeout  time.Time
}

// GitCommandService manages request/response correlation for structured Git commands.
type GitCommandService struct {
	pendingCommands sync.Map
	done   chan struct{}
	sender GitCommandSender
}

// NewGitCommandService creates a new git command service and wires response callbacks.
func NewGitCommandService(cm *RunnerConnectionManager) *GitCommandService {
	s := &GitCommandService{
		done: make(chan struct{}),
	}

	if cm != nil {
		cm.SetGitCommandResultCallback(func(runnerID int64, data *runnerv1.GitCommandResult) {
			s.CompleteCommand(data.GetRequestId(), runnerID, data)
		})
	}

	go s.cleanupLoop()
	return s
}

// Stop gracefully stops the git command service.
func (s *GitCommandService) Stop() {
	close(s.done)
}

// SetSender wires the command sender after gRPC initialization.
func (s *GitCommandService) SetSender(sender GitCommandSender) {
	s.sender = sender
}

func (s *GitCommandService) registerCommand(requestID string, timeout time.Duration) chan *runnerv1.GitCommandResult {
	resultCh := make(chan *runnerv1.GitCommandResult, 1)
	s.pendingCommands.Store(requestID, &pendingGitCommand{
		resultCh: resultCh,
		timeout:  time.Now().Add(timeout),
	})
	return resultCh
}

// CompleteCommand completes a pending git command with the async result from Runner.
func (s *GitCommandService) CompleteCommand(requestID string, _ int64, result *runnerv1.GitCommandResult) {
	if v, ok := s.pendingCommands.LoadAndDelete(requestID); ok {
		pending := v.(*pendingGitCommand)
		select {
		case pending.resultCh <- result:
		default:
		}
	}
}

// Execute sends a git command to a runner and waits for the async response.
func (s *GitCommandService) Execute(ctx context.Context, runnerID int64, cmd *runnerv1.GitCommand) (*runnerv1.GitCommandResult, error) {
	if s.sender == nil {
		return nil, ErrCommandSenderNotSet
	}

	requestID := cmd.GetRequestId()
	if requestID == "" {
		requestID = uuid.New().String()
		cmd.RequestId = requestID
	}

	resultCh := s.registerCommand(requestID, GitCommandTimeout)
	if err := s.sender.SendGitCommand(ctx, runnerID, cmd); err != nil {
		s.pendingCommands.Delete(requestID)
		return nil, err
	}

	select {
	case result := <-resultCh:
		return result, nil
	case <-ctx.Done():
		s.pendingCommands.Delete(requestID)
		return nil, ctx.Err()
	case <-time.After(GitCommandTimeout):
		s.pendingCommands.Delete(requestID)
		return &runnerv1.GitCommandResult{
			RequestId: requestID,
			PodKey:    cmd.GetPodKey(),
			Ok:        false,
			Code:      "command_timeout",
			Message:   "git command timed out",
		}, nil
	}
}

func (s *GitCommandService) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			now := time.Now()
			s.pendingCommands.Range(func(key, value any) bool {
				pending := value.(*pendingGitCommand)
				if now.After(pending.timeout) {
					if v, ok := s.pendingCommands.LoadAndDelete(key); ok {
						expired := v.(*pendingGitCommand)
						select {
						case expired.resultCh <- &runnerv1.GitCommandResult{
							RequestId: key.(string),
							Ok:        false,
							Code:      "command_timeout",
							Message:   "git command timed out",
						}:
						default:
						}
					}
				}
				return true
			})
		}
	}
}
