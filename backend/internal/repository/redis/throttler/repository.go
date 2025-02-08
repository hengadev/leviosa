package throttlerRepository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const THROTTLERPREFIX = "throttler:"

type Repository struct {
	client *redis.Client
}

func New(ctx context.Context, client *redis.Client) *Repository {
	return &Repository{client}
}

func (r *Repository) GetClient() *redis.Client {
	return r.client
}
