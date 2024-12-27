package otpRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/otp"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/redis/go-redis/v9"
)

func (o *Repository) getExistingOTP(ctx context.Context, key string) (*otpService.OTP, error) {
	data, err := o.client.Get(ctx, key).Bytes()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return nil, rp.NewNotFoundErr(err, "existing OTP with specified emailHash")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextErr(err)
		default:
			return nil, rp.NewDatabaseErr(fmt.Errorf("failed to check existing OTP: %w", err))
		}
	}

	var existingData otpService.OTP
	if err := json.Unmarshal(data, &existingData); err != nil {
		return nil, rp.NewDatabaseErr(fmt.Errorf("failed to parse existing OTP data: %w", err))
	}

	return &existingData, nil
}
