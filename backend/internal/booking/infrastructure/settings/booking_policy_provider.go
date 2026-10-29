// Package settings adapts the settings domain to the booking module's
// BookingPolicyProvider port (in-process, modular monolith).
package settings

import (
	"context"
	"fmt"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	bookingPorts "github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	settingsPorts "github.com/Leviosa-care/leviosa/backend/internal/settings/ports"
)

// InProcessBookingPolicyProvider reads the booking policy from the settings
// service on every call: an administrator's change applies to the next slot
// query, with no cache to invalidate.
type InProcessBookingPolicyProvider struct {
	settings settingsPorts.SettingsService
}

func NewInProcessBookingPolicyProvider(settings settingsPorts.SettingsService) bookingPorts.BookingPolicyProvider {
	return &InProcessBookingPolicyProvider{settings: settings}
}

func (p *InProcessBookingPolicyProvider) GetBookingPolicy(ctx context.Context) (domain.BookingPolicy, error) {
	policy, err := p.settings.GetBookingPolicy(ctx)
	if err != nil {
		return domain.BookingPolicy{}, fmt.Errorf("get booking policy from settings: %w", err)
	}
	return domain.BookingPolicy{
		MinimumNotice: time.Duration(policy.MinimumNoticeHours) * time.Hour,
		Horizon:       time.Duration(policy.HorizonWeeks) * 7 * 24 * time.Hour,
	}, nil
}
