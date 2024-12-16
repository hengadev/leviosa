package otpRepository

import (
	"context"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

func (o *Repository) InvalidateOTP(ctx context.Context, email string) error {
	err := o.client.Del(ctx, OTPPREFIX+email).Err()
	switch {
	case err == redis.Nil:
		return rp.NewNotFoundError(err)
	case err != nil:
		return rp.NewRessourceDeleteErr(err)
	}
	return nil
}
