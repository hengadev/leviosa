// Package settings reads the settings module in process for authuser.
package settings

import (
	"context"
	"fmt"
	"time"

	authuserPorts "github.com/Leviosa-care/leviosa/backend/internal/authuser/ports"
	settingsPorts "github.com/Leviosa-care/leviosa/backend/internal/settings/ports"
)

// InProcessOTPSettingsProvider reads the verification code settings from the
// settings service on every call, with no cache: a change made by an
// administrator applies to the next code.
type InProcessOTPSettingsProvider struct {
	settings settingsPorts.SettingsService
}

// NewInProcessOTPSettingsProvider creates the provider.
func NewInProcessOTPSettingsProvider(settings settingsPorts.SettingsService) authuserPorts.OTPSettingsProvider {
	return &InProcessOTPSettingsProvider{settings: settings}
}

// GetOTPDuration returns the "otp_duration" setting, stored in seconds.
func (p *InProcessOTPSettingsProvider) GetOTPDuration(ctx context.Context) (time.Duration, error) {
	resp, err := p.settings.GetOTPDuration(ctx)
	if err != nil {
		return 0, fmt.Errorf("get OTP duration from settings: %w", err)
	}
	return time.Duration(resp.Duration) * time.Second, nil
}
