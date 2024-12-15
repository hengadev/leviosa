package otpService

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) CreateOTP(ctx context.Context, email string) (*OTP, error) {
	otp, err := NewOTP(email)
	if err != nil {
		return nil, fmt.Errorf("create OTP instance: %w", err)
	}
	err = s.Repo.StoreOTP(ctx, otp)
	switch {
	case errors.Is(err, repository.ErrDatabase):
	case err != nil:
		return nil, fmt.Errorf("")
	}

	return otp, nil
}
