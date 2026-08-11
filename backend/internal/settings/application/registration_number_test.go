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

func TestGetCompanyRegistrationNumber(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the stored registration number", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("GetString", ctx, settings.CompanyRegistrationNumber).
			Return(&domain.Setting[string]{Key: settings.CompanyRegistrationNumber, Value: "12345678901234"}, nil)

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{})

		resp, err := svc.GetCompanyRegistrationNumber(ctx)
		require.NoError(t, err)
		assert.Equal(t, "12345678901234", resp.RegistrationNumber)
		repo.AssertExpectations(t)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("GetString", ctx, settings.CompanyRegistrationNumber).
			Return(nil, errors.New("not found"))

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{})

		_, err := svc.GetCompanyRegistrationNumber(ctx)
		assert.Error(t, err)
	})
}

func TestSetCompanyRegistrationNumber(t *testing.T) {
	ctx := context.Background()

	t.Run("persists a valid SIRET (14 digits) and publishes an update", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("SetString", ctx, mock.MatchedBy(func(s *domain.Setting[string]) bool {
			return s.Key == settings.CompanyRegistrationNumber && s.Value == "12345678901234"
		})).Return(nil)

		publisher := &mockEventPublisher{}
		publisher.On("PublishSettingUpdate", ctx, settings.CompanyRegistrationNumber, "12345678901234").Return(nil)

		svc := settingsApp.New(repo, nil, nil, publisher)

		resp, err := svc.SetCompanyRegistrationNumber(ctx, &domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: "12345678901234"})
		require.NoError(t, err)
		assert.True(t, resp.Success)
		repo.AssertExpectations(t)
		publisher.AssertExpectations(t)
	})

	t.Run("persists a valid SIREN (9 digits)", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("SetString", ctx, mock.Anything).Return(nil)
		publisher := &mockEventPublisher{}
		publisher.On("PublishSettingUpdate", ctx, settings.CompanyRegistrationNumber, "123456789").Return(nil)

		svc := settingsApp.New(repo, nil, nil, publisher)
		resp, err := svc.SetCompanyRegistrationNumber(ctx, &domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: "123456789"})
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("trims surrounding whitespace before persisting", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("SetString", ctx, mock.MatchedBy(func(s *domain.Setting[string]) bool {
			return s.Value == "123456789"
		})).Return(nil)
		publisher := &mockEventPublisher{}
		publisher.On("PublishSettingUpdate", ctx, settings.CompanyRegistrationNumber, "123456789").Return(nil)

		svc := settingsApp.New(repo, nil, nil, publisher)
		resp, err := svc.SetCompanyRegistrationNumber(ctx, &domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: "  123456789  "})
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("rejects an empty registration number", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{})

		_, err := svc.SetCompanyRegistrationNumber(ctx, &domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: ""})
		assert.Error(t, err)
		repo.AssertNotCalled(t, "SetString", mock.Anything, mock.Anything)
	})

	t.Run("rejects a registration number containing letters", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{})

		_, err := svc.SetCompanyRegistrationNumber(ctx, &domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: "1234ABCD9012"})
		assert.Error(t, err)
		repo.AssertNotCalled(t, "SetString", mock.Anything, mock.Anything)
	})

	t.Run("rejects a registration number of the wrong length", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{})

		_, err := svc.SetCompanyRegistrationNumber(ctx, &domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: "12345"})
		assert.Error(t, err)
		repo.AssertNotCalled(t, "SetString", mock.Anything, mock.Anything)
	})

	t.Run("does not publish when persistence fails", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("SetString", ctx, mock.Anything).Return(errors.New("db error"))
		publisher := &mockEventPublisher{}

		svc := settingsApp.New(repo, nil, nil, publisher)
		_, err := svc.SetCompanyRegistrationNumber(ctx, &domain.SetCompanyRegistrationNumberRequest{RegistrationNumber: "123456789"})
		assert.Error(t, err)
		publisher.AssertNotCalled(t, "PublishSettingUpdate", mock.Anything, mock.Anything, mock.Anything)
	})
}
