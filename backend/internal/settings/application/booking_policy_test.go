package settings_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	settingsApp "github.com/Leviosa-care/leviosa/backend/internal/settings/application"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetBookingPolicy(t *testing.T) {
	ctx := context.Background()
	notFound := errs.NewRepositoryNotFoundErr(errors.New("no rows"), "int setting")

	t.Run("answers the defaults while unset", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("GetInt", ctx, settings.BookingMinimumNoticeHours).Return(nil, notFound)
		repo.On("GetInt", ctx, settings.BookingHorizonWeeks).Return(nil, notFound)
		repo.On("GetInt", ctx, settings.BookingPaymentHoldMinutes).Return(nil, notFound)

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		resp, err := svc.GetBookingPolicy(ctx)
		require.NoError(t, err)
		assert.Equal(t, 12, resp.MinimumNoticeHours)
		assert.Equal(t, 8, resp.HorizonWeeks)
		assert.Equal(t, 15, resp.PaymentHoldMinutes)
	})

	t.Run("answers the stored values", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("GetInt", ctx, settings.BookingMinimumNoticeHours).
			Return(&domain.Setting[int]{Key: settings.BookingMinimumNoticeHours, Value: 0}, nil)
		repo.On("GetInt", ctx, settings.BookingHorizonWeeks).
			Return(&domain.Setting[int]{Key: settings.BookingHorizonWeeks, Value: 4}, nil)
		repo.On("GetInt", ctx, settings.BookingPaymentHoldMinutes).
			Return(&domain.Setting[int]{Key: settings.BookingPaymentHoldMinutes, Value: 30}, nil)

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		resp, err := svc.GetBookingPolicy(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, resp.MinimumNoticeHours)
		assert.Equal(t, 4, resp.HorizonWeeks)
		assert.Equal(t, 30, resp.PaymentHoldMinutes)
	})

	t.Run("propagates a repository failure", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("GetInt", ctx, settings.BookingMinimumNoticeHours).Return(nil, errors.New("connection refused"))

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		_, err := svc.GetBookingPolicy(ctx)
		assert.Error(t, err)
	})
}

func TestSetBookingPolicy(t *testing.T) {
	ctx := context.Background()

	t.Run("persists the three values", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("SetInt", ctx, mock.MatchedBy(func(s *domain.Setting[int]) bool {
			return s.Key == settings.BookingMinimumNoticeHours && s.Value == 24
		})).Return(nil)
		repo.On("SetInt", ctx, mock.MatchedBy(func(s *domain.Setting[int]) bool {
			return s.Key == settings.BookingHorizonWeeks && s.Value == 12
		})).Return(nil)
		repo.On("SetInt", ctx, mock.MatchedBy(func(s *domain.Setting[int]) bool {
			return s.Key == settings.BookingPaymentHoldMinutes && s.Value == 20
		})).Return(nil)

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		resp, err := svc.SetBookingPolicy(ctx, &domain.SetBookingPolicyRequest{MinimumNoticeHours: 24, HorizonWeeks: 12, PaymentHoldMinutes: 20})
		require.NoError(t, err)
		assert.True(t, resp.Success)
		repo.AssertExpectations(t)
	})

	for name, request := range map[string]domain.SetBookingPolicyRequest{
		"negative notice":       {MinimumNoticeHours: -1, HorizonWeeks: 8, PaymentHoldMinutes: 15},
		"notice over a week":    {MinimumNoticeHours: 169, HorizonWeeks: 8, PaymentHoldMinutes: 15},
		"zero horizon":          {MinimumNoticeHours: 12, HorizonWeeks: 0, PaymentHoldMinutes: 15},
		"horizon over a year":   {MinimumNoticeHours: 12, HorizonWeeks: 53, PaymentHoldMinutes: 15},
		"missing payment hold":  {MinimumNoticeHours: 12, HorizonWeeks: 8},
		"payment hold over 1 h": {MinimumNoticeHours: 12, HorizonWeeks: 8, PaymentHoldMinutes: 61},
	} {
		t.Run("rejects "+name, func(t *testing.T) {
			repo := &mockSettingsRepository{}
			svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

			_, err := svc.SetBookingPolicy(ctx, &request)
			assert.ErrorIs(t, err, errs.ErrInvalidValue)
			repo.AssertNotCalled(t, "SetInt", mock.Anything, mock.Anything)
		})
	}
}
