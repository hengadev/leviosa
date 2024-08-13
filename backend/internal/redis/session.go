package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	Client *redis.Client
}

func NewSessionRepository(ctx context.Context, client *redis.Client) *SessionRepository {
	return &SessionRepository{client}
}
