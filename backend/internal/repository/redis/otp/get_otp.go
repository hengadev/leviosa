package otpRepository

import (
	"context"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

func (o *Repository) GetOTP(ctx context.Context, email string) ([]byte, error) {
	val, err := o.client.Get(ctx, OTPPREFIX+email).Bytes()
	if err != nil {
		switch {
		case err == redis.Nil:
			return nil, rp.NewNotFoundError(err, "OTP key")
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	return val, nil
}
