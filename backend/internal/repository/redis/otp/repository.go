package otpRepository

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	otpKeyFormat = "otp:verification:%s"
)

type Repository struct {
	client *redis.Client
}

func New(ctx context.Context, client *redis.Client) *Repository {
	return &Repository{client}
}

func getOTPKey(emailHash string) string {
	return fmt.Sprintf(otpKeyFormat, emailHash)
}
