package otpRepository

import (
	"context"
	"encoding/json"
	"errors"
	"net"

	"github.com/GaryHY/event-reservation-app/internal/domain/otp"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

// TODO: do better error handling for that brother

func (o *Repository) StoreOTP(ctx context.Context, otp *otpService.OTP) error {
	// check if otp exists
	// forget the value got, we do not care
	_, err := o.client.Get(ctx, OTPPREFIX+otp.Email).Bytes()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return rp.NewNotFoundError(err, "OTP key")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextError(err)
		default:
			return rp.NewDatabaseErr(err)
		}

	}
	encoded, err := json.Marshal(otp)
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	err = o.client.Set(ctx, OTPPREFIX+otp.Email, encoded, otpService.OTPDURATION).Err()
	if err != nil {
		switch {
		case errors.Is(err, redis.ErrClosed), errors.As(err, &net.OpError{}):
			return rp.NewDatabaseErr(err)
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextError(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}
	return nil
}
