package sessionRepository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const SESSIONPREFIX = "session:"

type Repository struct {
	client *redis.Client
}

func New(ctx context.Context, client *redis.Client) *Repository {
	return &Repository{client}
}

func (r *Repository) GetClient() *redis.Client {
	return r.client
}
