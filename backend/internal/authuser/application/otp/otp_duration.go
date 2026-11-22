package otp

import (
	"context"
	"log/slog"
	"time"
)

// otpDuration is how long a new verification code stays valid: the duration
// set in the admin settings, or defaultOTPDuration when there is no settings
// provider, the setting can't be read, or it is outside the range the
// settings page accepts (1 minute to 1 hour).
func (s *OTPService) otpDuration(ctx context.Context) time.Duration {
	fallback := time.Duration(defaultOTPDuration) * time.Minute
	if s.settings == nil {
		return fallback
	}
	d, err := s.settings.GetOTPDuration(ctx)
	if err != nil {
		slog.WarnContext(ctx, "read OTP duration setting, using the default", "error", err)
		return fallback
	}
	if d < time.Duration(minOTPDuration)*time.Minute || d > time.Duration(maxOTPDuration)*time.Minute {
		slog.WarnContext(ctx, "OTP duration setting out of range, using the default", "seconds", int(d/time.Second))
		return fallback
	}
	return d
}
