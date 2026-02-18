package apikey

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Service implements the API key business logic
type Service struct {
	db          *gorm.DB
	redisClient *redis.Client
}

// Compile-time interface check
var _ Interface = (*Service)(nil)

// NewService creates a new API key service
func NewService(db *gorm.DB, redisClient *redis.Client) *Service {
	return &Service{
		db:          db,
		redisClient: redisClient,
	}
}
