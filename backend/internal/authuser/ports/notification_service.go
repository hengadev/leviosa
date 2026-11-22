package ports

import (
	"context"
	"time"
)

// NotificationService defines the interface for sending OTP notifications.
// This is an anti-corruption port — authuser does not depend on the notification
// module's concrete types.
type NotificationService interface {
	// SendOTPEmail sends a verification code valid for validFor.
	SendOTPEmail(ctx context.Context, email, otp string, validFor time.Duration) error
}
