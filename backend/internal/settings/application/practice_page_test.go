package settings_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	settingsApp "github.com/Leviosa-care/leviosa/backend/internal/settings/application"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetPracticePagePartner(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the stored partner ID", func(t *testing.T) {
		partnerID := uuid.NewString()
		repo := &mockSettingsRepository{}
		repo.On("GetString", ctx, settings.PracticePagePartnerID).
			Return(&domain.Setting[string]{Key: settings.PracticePagePartnerID, Value: partnerID}, nil)

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		resp, err := svc.GetPracticePagePartner(ctx)
		require.NoError(t, err)
		assert.Equal(t, partnerID, resp.PartnerID)
	})

	t.Run("propagates the not-found error when unset", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("GetString", ctx, settings.PracticePagePartnerID).
			Return(nil, errors.New("record not found"))

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		_, err := svc.GetPracticePagePartner(ctx)
		assert.Error(t, err)
	})
}

func TestSetPracticePagePartner(t *testing.T) {
	ctx := context.Background()

	t.Run("persists a valid partner ID", func(t *testing.T) {
		partnerID := uuid.NewString()
		repo := &mockSettingsRepository{}
		repo.On("SetString", ctx, mock.MatchedBy(func(s *domain.Setting[string]) bool {
			return s.Key == settings.PracticePagePartnerID && s.Value == partnerID
		})).Return(nil)

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		resp, err := svc.SetPracticePagePartner(ctx, &domain.SetPracticePagePartnerRequest{PartnerID: partnerID})
		require.NoError(t, err)
		assert.True(t, resp.Success)
		repo.AssertExpectations(t)
	})

	t.Run("rejects a non-UUID partner ID", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		_, err := svc.SetPracticePagePartner(ctx, &domain.SetPracticePagePartnerRequest{PartnerID: "not-a-uuid"})
		assert.Error(t, err)
		repo.AssertNotCalled(t, "SetString", mock.Anything, mock.Anything)
	})

	t.Run("rejects an empty partner ID", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		_, err := svc.SetPracticePagePartner(ctx, &domain.SetPracticePagePartnerRequest{PartnerID: ""})
		assert.Error(t, err)
		repo.AssertNotCalled(t, "SetString", mock.Anything, mock.Anything)
	})
}

func TestGetPracticePageProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the stored product ID", func(t *testing.T) {
		productID := uuid.NewString()
		repo := &mockSettingsRepository{}
		repo.On("GetString", ctx, settings.PracticePageProductID).
			Return(&domain.Setting[string]{Key: settings.PracticePageProductID, Value: productID}, nil)

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		resp, err := svc.GetPracticePageProduct(ctx)
		require.NoError(t, err)
		assert.Equal(t, productID, resp.ProductID)
	})
}

func TestSetPracticePageProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("persists a valid product ID", func(t *testing.T) {
		productID := uuid.NewString()
		repo := &mockSettingsRepository{}
		repo.On("SetString", ctx, mock.MatchedBy(func(s *domain.Setting[string]) bool {
			return s.Key == settings.PracticePageProductID && s.Value == productID
		})).Return(nil)

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		resp, err := svc.SetPracticePageProduct(ctx, &domain.SetPracticePageProductRequest{ProductID: productID})
		require.NoError(t, err)
		assert.True(t, resp.Success)
		repo.AssertExpectations(t)
	})

	t.Run("rejects a non-UUID product ID", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		_, err := svc.SetPracticePageProduct(ctx, &domain.SetPracticePageProductRequest{ProductID: "42"})
		assert.Error(t, err)
		repo.AssertNotCalled(t, "SetString", mock.Anything, mock.Anything)
	})
}

func TestSetPracticePageVideo(t *testing.T) {
	ctx := context.Background()

	t.Run("persists a valid video URL", func(t *testing.T) {
		videoURL := "https://media.leviosa.care/practice.mp4"
		repo := &mockSettingsRepository{}
		repo.On("GetString", ctx, settings.PracticePageVideoURL).Return(nil, errNotSet)
		repo.On("SetString", ctx, mock.MatchedBy(func(s *domain.Setting[string]) bool {
			return s.Key == settings.PracticePageVideoURL && s.Value == videoURL
		})).Return(nil)

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		resp, err := svc.SetPracticePageVideo(ctx, &domain.SetPracticePageVideoRequest{VideoURL: videoURL})
		require.NoError(t, err)
		assert.True(t, resp.Success)
		repo.AssertExpectations(t)
	})

	t.Run("accepts an empty URL, clearing the video", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		repo.On("GetString", ctx, settings.PracticePageVideoURL).Return(nil, errNotSet)
		repo.On("SetString", ctx, mock.MatchedBy(func(s *domain.Setting[string]) bool {
			return s.Key == settings.PracticePageVideoURL && s.Value == ""
		})).Return(nil)

		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		resp, err := svc.SetPracticePageVideo(ctx, &domain.SetPracticePageVideoRequest{VideoURL: ""})
		require.NoError(t, err)
		assert.True(t, resp.Success)
		repo.AssertExpectations(t)
	})

	t.Run("rejects a non-http URL", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		_, err := svc.SetPracticePageVideo(ctx, &domain.SetPracticePageVideoRequest{VideoURL: "ftp://example.com/video.mp4"})
		assert.Error(t, err)
		repo.AssertNotCalled(t, "SetString", mock.Anything, mock.Anything)
	})

	t.Run("rejects a URL without a host", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		_, err := svc.SetPracticePageVideo(ctx, &domain.SetPracticePageVideoRequest{VideoURL: "/video.mp4"})
		assert.Error(t, err)
		repo.AssertNotCalled(t, "SetString", mock.Anything, mock.Anything)
	})

	t.Run("rejects a URL exceeding 500 characters", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		svc := settingsApp.New(repo, nil, nil, &mockEventPublisher{}, nil)

		_, err := svc.SetPracticePageVideo(ctx, &domain.SetPracticePageVideoRequest{VideoURL: "https://example.com/" + strings.Repeat("a", 500)})
		assert.Error(t, err)
		repo.AssertNotCalled(t, "SetString", mock.Anything, mock.Anything)
	})
}
