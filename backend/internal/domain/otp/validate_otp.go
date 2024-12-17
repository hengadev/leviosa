package otpService

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) ValidateOTP(ctx context.Context, otp *OTP) error {
	storedOTP, err := s.Repo.GetOTP(ctx, otp.Email)
	switch {
	case errors.Is(err, rp.ErrNotFound):
		return fmt.Errorf("no OTP found for given email %s", otp.Email)
	case errors.Is(err, rp.ErrExpiredToken):
		return fmt.Errorf("the OTP for email %s has expired; please request a new one", otp.Email)
	case err != nil:
		return fmt.Errorf("failed to retrieve OTP: %w", err)
	}
	if otp.Value != storedOTP.Value {
		return fmt.Errorf("invalid OTP provided")
	}
	return nil
}
