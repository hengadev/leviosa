package throttlerRepository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const THROTTLERPREFIX = "throttler:"

type Repository struct {
	Client *redis.Client
}

func New(ctx context.Context, client *redis.Client) *Repository {
	return &Repository{client}
}
