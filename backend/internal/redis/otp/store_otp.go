package otpRepository

import (
	"context"
	"encoding/json"

	"github.com/GaryHY/event-reservation-app/internal/domain/otp"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

func (o *Repository) StoreOTP(ctx context.Context, otp *otpService.OTP) error {
	// check if otp exists
	// forget the value got, we do not care
	_, err := o.Client.Get(ctx, OTPPREFIX+otp.Email).Bytes()
	if err != redis.Nil {
		return rp.NewDatabaseErr(err)
	}
	encoded, err := json.Marshal(otp)
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	err = o.Client.Set(ctx, OTPPREFIX+otp.Email, encoded, otpService.OTPDURATION).Err()
	if err != nil {
		return rp.NewRessourceUpdateErr(err)
	}
	return nil
}
