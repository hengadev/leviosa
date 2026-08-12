package settings_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	settingsApp "github.com/Leviosa-care/leviosa/backend/internal/settings/application"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetCompanyHostingProvider(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the stored hosting provider identity", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("GetString", ctx, settings.CompanyHostingProvider).
			Return(&domain.Setting[string]{Key: settings.CompanyHostingProvider, Value: "OVH SAS, 2 rue Kellermann, 59100 Roubaix, France"}, nil)

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		resp, err := svc.GetCompanyHostingProvider(ctx)
		require.NoError(t, err)
		assert.Equal(t, "OVH SAS, 2 rue Kellermann, 59100 Roubaix, France", resp.HostingProvider)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("GetString", ctx, settings.CompanyHostingProvider).
			Return(nil, errors.New("not found"))

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		_, err := svc.GetCompanyHostingProvider(ctx)
		assert.Error(t, err)
	})
}

func TestSetCompanyHostingProvider(t *testing.T) {
	ctx := context.Background()

	t.Run("persists a valid hosting provider identity and publishes an update", func(t *testing.T) {
		value := "OVH SAS, 2 rue Kellermann, 59100 Roubaix, France"
		repo := &mockSettingsRepository{}
		repo.On("SetString", ctx, mock.MatchedBy(func(s *domain.Setting[string]) bool {
			return s.Key == settings.CompanyHostingProvider && s.Value == value
		})).Return(nil)

		publisher := &mockEventPublisher{}
		publisher.On("PublishSettingUpdate", ctx, settings.CompanyHostingProvider, value).Return(nil)

		svc := settingsApp.New(repo, nil, nil, publisher, nil)

		resp, err := svc.SetCompanyHostingProvider(ctx, &domain.SetCompanyHostingProviderRequest{HostingProvider: value})
		require.NoError(t, err)
		assert.True(t, resp.Success)
		repo.AssertExpectations(t)
		publisher.AssertExpectations(t)
	})

	t.Run("rejects an empty hosting provider identity", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		_, err := svc.SetCompanyHostingProvider(ctx, &domain.SetCompanyHostingProviderRequest{HostingProvider: ""})
		assert.Error(t, err)
		repo.AssertNotCalled(t, "SetString", mock.Anything, mock.Anything)
	})

	t.Run("rejects a whitespace-only hosting provider identity", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		_, err := svc.SetCompanyHostingProvider(ctx, &domain.SetCompanyHostingProviderRequest{HostingProvider: "   \n\t   "})
		assert.Error(t, err)
		repo.AssertNotCalled(t, "SetString", mock.Anything, mock.Anything)
	})

	t.Run("rejects a hosting provider identity exceeding 500 characters", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		_, err := svc.SetCompanyHostingProvider(ctx, &domain.SetCompanyHostingProviderRequest{HostingProvider: strings.Repeat("A", 501)})
		assert.Error(t, err)
		repo.AssertNotCalled(t, "SetString", mock.Anything, mock.Anything)
	})

	t.Run("does not publish when persistence fails", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("SetString", ctx, mock.Anything).Return(errors.New("db error"))
		publisher := &mockEventPublisher{}

		svc := settingsApp.New(repo, nil, nil, publisher, nil)
		_, err := svc.SetCompanyHostingProvider(ctx, &domain.SetCompanyHostingProviderRequest{HostingProvider: "OVH SAS, Roubaix, France"})
		assert.Error(t, err)
		publisher.AssertNotCalled(t, "PublishSettingUpdate", mock.Anything, mock.Anything, mock.Anything)
	})
}
