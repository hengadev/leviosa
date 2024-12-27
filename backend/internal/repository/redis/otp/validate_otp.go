package otpRepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	otpService "github.com/GaryHY/event-reservation-app/internal/domain/otp"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// Helper function for validating OTP
func (o *Repository) ValidateOTP(ctx context.Context, emailHash, providedOTP string) error {
	key := fmt.Sprintf(otpKeyFormat, emailHash)
	otpData, err := o.getExistingOTP(ctx, key)
	if err != nil {
		return err
	}

	if otpData == nil {
		return rp.NewNotFoundErr(errors.New("OTP not found or expired"), "otp")
	}

	if time.Now().After(otpData.ExpiresAt) {
		return rp.NewValidationErr(errors.New("expired OTP"), "OTP")
	}

	// Check attempts
	if otpData.Attempts >= otpService.MaxOTPAttempts {
		// Delete expired OTP
		o.client.Del(ctx, key)
		return rp.NewValidationErr(errors.New("max attempts exceeded"), "OTP")
	}

	// Increment attempts
	otpData.Attempts++
	if err := o.setOTP(ctx, key, otpData); err != nil {
		return err
	}

	// Validate OTP
	if providedOTP != otpData.Code {
		return rp.NewValidationErr(errors.New("provided OTP code does not match stored OTP code"), "OTP")
	}

	// Delete OTP after successful validation
	o.client.Del(ctx, key)
	return nil
}
