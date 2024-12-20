package otpRepository

import (
	"context"
	"errors"
	"fmt"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

func (o *Repository) InvalidateOTP(ctx context.Context, email string) error {
	result := o.client.Del(ctx, OTPPREFIX+email)

	if err := result.Err(); err != nil {
		switch {
		case errors.Is(err, redis.ErrClosed):
			return rp.NewDatabaseErr(err)
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextError(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}

	if result.Val() == 0 {
		return rp.NewNotFoundError(fmt.Errorf("key does not exist"), "OTP")
	}
	return nil
}
