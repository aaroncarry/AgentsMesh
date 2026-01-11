package runner

import (
	"sync"
	"time"

	"github.com/anthropics/agentmesh/backend/internal/domain/runner"
	"gorm.io/gorm"
)

// Service handles runner operations
type Service struct {
	db            *gorm.DB
	activeRunners sync.Map // map[runnerID]*ActiveRunner
}

// ActiveRunner represents an active runner connection
type ActiveRunner struct {
	Runner   *runner.Runner
	LastPing time.Time
	PodCount int
}

// NewService creates a new runner service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}
