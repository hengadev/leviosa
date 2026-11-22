package ports

import (
	"context"
	"time"
)

// OTPSettingsProvider reads the verification code settings an administrator
// sets in the admin settings page.
type OTPSettingsProvider interface {
	// GetOTPDuration is how long a new verification code stays valid.
	GetOTPDuration(ctx context.Context) (time.Duration, error)
}
