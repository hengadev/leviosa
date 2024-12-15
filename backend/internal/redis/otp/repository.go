package otpRepository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const OTPPREFIX = "otp:"

type Repository struct {
	Client *redis.Client
}

func New(ctx context.Context, client *redis.Client) *Repository {
	return &Repository{client}
}
