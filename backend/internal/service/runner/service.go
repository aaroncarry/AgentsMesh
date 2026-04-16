package runner

import (
	"context"
	"sync"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/anthropics/agentsmesh/backend/internal/service/billing"
)

// GrantQuerier queries resource grants. Optional dependency for visibility checks.
type GrantQuerier interface {
	GetGrantedResourceIDs(ctx context.Context, resourceType string, userID int64, orgID int64) ([]string, error)
}

// Service handles runner operations
type Service struct {
	repo           runner.RunnerRepository
	billingService *billing.Service
	grantQuerier   GrantQuerier
	activeRunners  sync.Map // map[runnerID]*ActiveRunner
}

// ActiveRunner represents an active runner connection
type ActiveRunner struct {
	Runner   *runner.Runner
	LastPing time.Time
	PodCount int
}

// NewService creates a new runner service
// billingService is optional - pass nil to skip quota checks (useful for tests)
func NewService(repo runner.RunnerRepository, billingService ...*billing.Service) *Service {
	s := &Service{
		repo: repo,
	}
	if len(billingService) > 0 {
		s.billingService = billingService[0]
	}
	return s
}

// SetGrantQuerier sets the optional grant querier for visibility checks.
func (s *Service) SetGrantQuerier(gq GrantQuerier) {
	s.grantQuerier = gq
}
