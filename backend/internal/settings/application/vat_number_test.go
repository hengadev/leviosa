package settings_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	settingsApp "github.com/Leviosa-care/leviosa/backend/internal/settings/application"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetCompanyVatNumber(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the stored VAT number", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("GetString", ctx, settings.CompanyVatNumber).
			Return(&domain.Setting[string]{Key: settings.CompanyVatNumber, Value: "FR12345678901"}, nil)

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		resp, err := svc.GetCompanyVatNumber(ctx)
		require.NoError(t, err)
		assert.Equal(t, "FR12345678901", resp.VatNumber)
	})

	t.Run("returns an empty VAT number for franchise-regime businesses", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("GetString", ctx, settings.CompanyVatNumber).
			Return(&domain.Setting[string]{Key: settings.CompanyVatNumber, Value: ""}, nil)

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		resp, err := svc.GetCompanyVatNumber(ctx)
		require.NoError(t, err)
		assert.Equal(t, "", resp.VatNumber)
	})
}

func TestSetCompanyVatNumber(t *testing.T) {
	ctx := context.Background()

	t.Run("persists a valid VAT number and publishes an update", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("SetString", ctx, mock.MatchedBy(func(s *domain.Setting[string]) bool {
			return s.Key == settings.CompanyVatNumber && s.Value == "FR12345678901"
		})).Return(nil)

		publisher := &mockEventPublisher{}
		publisher.On("PublishSettingUpdate", ctx, settings.CompanyVatNumber, "FR12345678901").Return(nil)

		svc := settingsApp.New(repo, nil, nil, publisher, nil)

		resp, err := svc.SetCompanyVatNumber(ctx, &domain.SetCompanyVatNumberRequest{VatNumber: "FR12345678901"})
		require.NoError(t, err)
		assert.True(t, resp.Success)
		repo.AssertExpectations(t)
		publisher.AssertExpectations(t)
	})

	// This is the core acceptance criterion for this field: a business under
	// the French "franchise en base de TVA" regime has no VAT number, and
	// submitting a blank value must succeed rather than error.
	t.Run("accepts an empty VAT number without error", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("SetString", ctx, mock.MatchedBy(func(s *domain.Setting[string]) bool {
			return s.Key == settings.CompanyVatNumber && s.Value == ""
		})).Return(nil)

		publisher := &mockEventPublisher{}
		publisher.On("PublishSettingUpdate", ctx, settings.CompanyVatNumber, "").Return(nil)

		svc := settingsApp.New(repo, nil, nil, publisher, nil)

		resp, err := svc.SetCompanyVatNumber(ctx, &domain.SetCompanyVatNumberRequest{VatNumber: ""})
		require.NoError(t, err)
		assert.True(t, resp.Success)
		repo.AssertExpectations(t)
		publisher.AssertExpectations(t)
	})

	t.Run("accepts a whitespace-only VAT number as blank", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("SetString", ctx, mock.MatchedBy(func(s *domain.Setting[string]) bool {
			return s.Value == ""
		})).Return(nil)
		publisher := &mockEventPublisher{}
		publisher.On("PublishSettingUpdate", ctx, settings.CompanyVatNumber, "").Return(nil)

		svc := settingsApp.New(repo, nil, nil, publisher, nil)

		resp, err := svc.SetCompanyVatNumber(ctx, &domain.SetCompanyVatNumberRequest{VatNumber: "   "})
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("rejects a malformed non-empty VAT number", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		_, err := svc.SetCompanyVatNumber(ctx, &domain.SetCompanyVatNumberRequest{VatNumber: "1"})
		assert.Error(t, err)
		repo.AssertNotCalled(t, "SetString", mock.Anything, mock.Anything)
	})

	t.Run("rejects a VAT number exceeding the max length", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		_, err := svc.SetCompanyVatNumber(ctx, &domain.SetCompanyVatNumberRequest{VatNumber: "FR123456789012345678901"})
		assert.Error(t, err)
		repo.AssertNotCalled(t, "SetString", mock.Anything, mock.Anything)
	})

	t.Run("does not publish when persistence fails", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("SetString", ctx, mock.Anything).Return(errors.New("db error"))
		publisher := &mockEventPublisher{}

		svc := settingsApp.New(repo, nil, nil, publisher, nil)
		_, err := svc.SetCompanyVatNumber(ctx, &domain.SetCompanyVatNumberRequest{VatNumber: "FR12345678901"})
		assert.Error(t, err)
		publisher.AssertNotCalled(t, "PublishSettingUpdate", mock.Anything, mock.Anything, mock.Anything)
	})
}
