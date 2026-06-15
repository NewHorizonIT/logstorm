package auth

import (
	"context"
	"time"

	"github.com/NewHorizonIT/logstorm/internal/infra/redis"
)

// Define repository layer for session management here
type sessionRepository struct {
	redisClient redis.Cache
}

// Initialize sessionRepository
func NewSessionRepository(redisClient redis.Cache) SessionRepository {
	return &sessionRepository{
		redisClient: redisClient,
	}
}

// StoreRefreshToken implements [SessionRepository].
func (s *sessionRepository) StoreRefreshToken(ctx context.Context, sid string, userID int, expiration time.Duration) error {
	redisKey := redis.SessionKey(sid)
	return s.redisClient.Set(ctx, redisKey, userID, expiration)
}

// DeleteSession implements [SessionRepository].
func (s *sessionRepository) DeleteSession(ctx context.Context, sid string) error {
	redisKey := redis.SessionKey(sid)
	return s.redisClient.Delete(ctx, redisKey)
}

// GetUserIDBySID implements [SessionRepository].
func (s *sessionRepository) GetUserIDBySID(ctx context.Context, sid string) (int, error) {
	redisKey := redis.SessionKey(sid)
	var userID int
	err := s.redisClient.Get(ctx, redisKey, &userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
