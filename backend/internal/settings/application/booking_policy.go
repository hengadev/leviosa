package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

// Booking policy settings (docs/issues/010-slot-engine.md): the minimum
// notice and the booking horizon the slot engine applies, and the payment
// hold an unpaid booking keeps its slot for (issue 014, ADR-0018). Unlike the practice
// page settings, the policy always has a value: an unset key reads as its
// default (12 h, 8 weeks, 15 min), so the public page and booking creation never wait
// on an administrator to configure it. Changing a value takes effect on the
// next slot query, without a release.

func (s *SettingsService) GetBookingPolicy(ctx context.Context) (*domain.GetBookingPolicyResponse, error) {
	notice, err := s.intOrDefault(ctx, settings.BookingMinimumNoticeHours, domain.DefaultBookingMinimumNoticeHours)
	if err != nil {
		return nil, fmt.Errorf("get booking minimum notice: %w", err)
	}
	horizon, err := s.intOrDefault(ctx, settings.BookingHorizonWeeks, domain.DefaultBookingHorizonWeeks)
	if err != nil {
		return nil, fmt.Errorf("get booking horizon: %w", err)
	}
	hold, err := s.intOrDefault(ctx, settings.BookingPaymentHoldMinutes, domain.DefaultPaymentHoldMinutes)
	if err != nil {
		return nil, fmt.Errorf("get booking payment hold: %w", err)
	}
	return &domain.GetBookingPolicyResponse{MinimumNoticeHours: notice, HorizonWeeks: horizon, PaymentHoldMinutes: hold}, nil
}

func (s *SettingsService) SetBookingPolicy(ctx context.Context, request *domain.SetBookingPolicyRequest) (*domain.SetBookingPolicyResponse, error) {
	if err := request.Valid(ctx); err != nil {
		return nil, errs.NewInvalidValueErr(err.Error())
	}

	if err := s.repo.SetInt(ctx, NewSetting(settings.BookingMinimumNoticeHours, request.MinimumNoticeHours)); err != nil {
		return nil, fmt.Errorf("set booking minimum notice: %w", err)
	}
	s.auditFieldUpdate(ctx, settings.BookingMinimumNoticeHours, request.MinimumNoticeHours)

	if err := s.repo.SetInt(ctx, NewSetting(settings.BookingHorizonWeeks, request.HorizonWeeks)); err != nil {
		return nil, fmt.Errorf("set booking horizon: %w", err)
	}
	s.auditFieldUpdate(ctx, settings.BookingHorizonWeeks, request.HorizonWeeks)

	if err := s.repo.SetInt(ctx, NewSetting(settings.BookingPaymentHoldMinutes, request.PaymentHoldMinutes)); err != nil {
		return nil, fmt.Errorf("set booking payment hold: %w", err)
	}
	s.auditFieldUpdate(ctx, settings.BookingPaymentHoldMinutes, request.PaymentHoldMinutes)

	return &domain.SetBookingPolicyResponse{Success: true}, nil
}

// intOrDefault reads an int setting, answering fallback when it is unset.
func (s *SettingsService) intOrDefault(ctx context.Context, key string, fallback int) (int, error) {
	setting, err := s.repo.GetInt(ctx, key)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return fallback, nil
		}
		return 0, err
	}
	return setting.Value, nil
}
