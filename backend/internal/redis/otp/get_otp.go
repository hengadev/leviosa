package otpRepository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/otp"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

func (o *Repository) GetOTP(ctx context.Context, email string) (*otpService.OTP, error) {
	var res otpService.OTP
	val, err := o.client.Get(ctx, OTPPREFIX+email).Bytes()
	switch {
	case err == redis.Nil:
		return nil, rp.NewNotFoundError(err)
	case err != nil:
		return nil, rp.NewDatabaseErr(err)
	}
	if err = json.Unmarshal(val, &res); err != nil {
		return nil, rp.NewInternalError(err)
	}
	if time.Now().After(res.ExpiresAt) {
		return nil, rp.NewExpiredTokenError("OTP", err)
	}
	return &res, nil
}
