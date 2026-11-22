package otp

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/authuser/ports"

	"github.com/hengadev/encx"
)

const (
	defaultOTPDuration    = 10 // Default OTP duration in minutes
	defaultOTPLength      = 6  // Default OTP length
	defaultOTPMaxAttempts = 3  // Default max attempts
)

type OTPService struct {
	repo            ports.OTPRepository
	crypto          encx.CryptoService
	cache           ports.OTPCache
	notificationSvc ports.NotificationService
	// settings reads the code duration set by an administrator; nil means
	// defaultOTPDuration.
	settings ports.OTPSettingsProvider
}

// Option configures the OTP service.
type Option func(*OTPService)

// WithSettingsProvider makes new codes last the duration set in the admin
// settings instead of defaultOTPDuration.
func WithSettingsProvider(p ports.OTPSettingsProvider) Option {
	return func(s *OTPService) { s.settings = p }
}

func New(ctx context.Context, repo ports.OTPRepository, crypto encx.CryptoService, notificationSvc ports.NotificationService, opts ...Option) (ports.OTPService, error) {
	// Initialize cache with hardcoded default values
	cache := domain.NewOTPCache(
		defaultOTPDuration,
		defaultOTPLength,
		defaultOTPMaxAttempts,
	)

	otpService := &OTPService{
		repo:            repo,
		crypto:          crypto,
		cache:           cache,
		notificationSvc: notificationSvc,
	}
	for _, opt := range opts {
		opt(otpService)
	}

	return otpService, nil
}
