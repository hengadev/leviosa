package settings_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	settingsApp "github.com/Leviosa-care/leviosa/backend/internal/settings/application"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errNotSet = errs.NewRepositoryNotFoundErr(errors.New("no rows"), "string setting")

// mp4Bytes starts like an MP4 file (an ftyp box), which is what the upload
// checks; the rest is filler.
func mp4Bytes() []byte {
	head := []byte("\x00\x00\x00\x18ftypmp42\x00\x00\x00\x00mp42isom")
	return append(head, bytes.Repeat([]byte{0}, 1024)...)
}

const previousUpload = "/media/practice-page/video/0b8a4a5e-6f0e-4f7b-9d55-6a9a0f3c1e11.mp4"

func videoSetting(value string) *domain.Setting[string] {
	return &domain.Setting[string]{Key: settings.PracticePageVideoURL, Value: value}
}

func TestUploadPracticePageVideo(t *testing.T) {
	ctx := context.Background()

	t.Run("stores an MP4 and points the setting at it", func(t *testing.T) {
		data := mp4Bytes()
		repo := &mockSettingsRepository{}
		media := &mockSettingsMedia{}
		var storedKey string
		media.On("PutObject", ctx, mock.MatchedBy(func(key string) bool {
			storedKey = key
			return domain.IsPracticePageVideoKey(key) && strings.HasSuffix(key, ".mp4")
		}), int64(len(data)), "video/mp4").Return(nil)
		repo.On("GetString", ctx, settings.PracticePageVideoURL).Return(nil, errNotSet)
		repo.On("SetString", ctx, mock.MatchedBy(func(s *domain.Setting[string]) bool {
			return s.Key == settings.PracticePageVideoURL && s.Value == "/media/"+storedKey
		})).Return(nil)

		svc := settingsApp.New(repo, media, nil, &mockEventPublisher{}, nil)
		resp, err := svc.UploadPracticePageVideo(ctx, bytes.NewReader(data), int64(len(data)))

		require.NoError(t, err)
		assert.Equal(t, "/media/"+storedKey, resp.VideoURL)
		assert.Equal(t, data, media.put[storedKey], "the whole file is stored, from its first byte")
		repo.AssertExpectations(t)
		media.AssertExpectations(t)
	})

	t.Run("deletes the video it replaces", func(t *testing.T) {
		data := mp4Bytes()
		repo := &mockSettingsRepository{}
		media := &mockSettingsMedia{}
		media.On("PutObject", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		media.On("DeleteObject", ctx, strings.TrimPrefix(previousUpload, "/media/")).Return(nil)
		repo.On("GetString", ctx, settings.PracticePageVideoURL).Return(videoSetting(previousUpload), nil)
		repo.On("SetString", ctx, mock.Anything).Return(nil)

		svc := settingsApp.New(repo, media, nil, &mockEventPublisher{}, nil)
		_, err := svc.UploadPracticePageVideo(ctx, bytes.NewReader(data), int64(len(data)))

		require.NoError(t, err)
		media.AssertExpectations(t)
	})

	t.Run("rejects a file that is not MP4 or WebM, whatever its name", func(t *testing.T) {
		data := []byte("\x89PNG\r\n\x1a\n" + strings.Repeat("x", 600))
		repo := &mockSettingsRepository{}
		media := &mockSettingsMedia{}

		svc := settingsApp.New(repo, media, nil, &mockEventPublisher{}, nil)
		_, err := svc.UploadPracticePageVideo(ctx, bytes.NewReader(data), int64(len(data)))

		require.ErrorIs(t, err, errs.ErrInvalidValue)
		media.AssertNotCalled(t, "PutObject", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		repo.AssertNotCalled(t, "SetString", mock.Anything, mock.Anything)
	})

	t.Run("rejects a file over 50 MB", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		media := &mockSettingsMedia{}

		svc := settingsApp.New(repo, media, nil, &mockEventPublisher{}, nil)
		_, err := svc.UploadPracticePageVideo(ctx, bytes.NewReader(mp4Bytes()), domain.PracticePageVideoMaxBytes+1)

		require.ErrorIs(t, err, errs.ErrInvalidValue)
		media.AssertNotCalled(t, "PutObject", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("removes the new object when the setting cannot be saved", func(t *testing.T) {
		data := mp4Bytes()
		repo := &mockSettingsRepository{}
		media := &mockSettingsMedia{}
		var storedKey string
		media.On("PutObject", ctx, mock.MatchedBy(func(key string) bool { storedKey = key; return true }), mock.Anything, mock.Anything).Return(nil)
		media.On("DeleteObject", ctx, mock.MatchedBy(func(key string) bool { return key == storedKey })).Return(nil)
		repo.On("GetString", ctx, settings.PracticePageVideoURL).Return(nil, errNotSet)
		repo.On("SetString", ctx, mock.Anything).Return(errors.New("database down"))

		svc := settingsApp.New(repo, media, nil, &mockEventPublisher{}, nil)
		_, err := svc.UploadPracticePageVideo(ctx, bytes.NewReader(data), int64(len(data)))

		require.Error(t, err)
		media.AssertExpectations(t)
	})
}

func TestSetPracticePageVideo_ReplacingAnUpload(t *testing.T) {
	ctx := context.Background()

	t.Run("clearing the video deletes the uploaded file", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		media := &mockSettingsMedia{}
		repo.On("GetString", ctx, settings.PracticePageVideoURL).Return(videoSetting(previousUpload), nil)
		repo.On("SetString", ctx, mock.MatchedBy(func(s *domain.Setting[string]) bool { return s.Value == "" })).Return(nil)
		media.On("DeleteObject", ctx, strings.TrimPrefix(previousUpload, "/media/")).Return(nil)

		svc := settingsApp.New(repo, media, nil, &mockEventPublisher{}, nil)
		_, err := svc.SetPracticePageVideo(ctx, &domain.SetPracticePageVideoRequest{VideoURL: ""})

		require.NoError(t, err)
		media.AssertExpectations(t)
	})

	t.Run("a failed delete still clears the video", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		media := &mockSettingsMedia{}
		repo.On("GetString", ctx, settings.PracticePageVideoURL).Return(videoSetting(previousUpload), nil)
		repo.On("SetString", ctx, mock.Anything).Return(nil)
		media.On("DeleteObject", ctx, mock.Anything).Return(errors.New("s3 down"))

		svc := settingsApp.New(repo, media, nil, &mockEventPublisher{}, nil)
		_, err := svc.SetPracticePageVideo(ctx, &domain.SetPracticePageVideoRequest{VideoURL: ""})

		require.NoError(t, err)
	})

	t.Run("replacing a link deletes nothing", func(t *testing.T) {
		repo := &mockSettingsRepository{}
		media := &mockSettingsMedia{}
		repo.On("GetString", ctx, settings.PracticePageVideoURL).Return(videoSetting("https://example.com/video.mp4"), nil)
		repo.On("SetString", ctx, mock.Anything).Return(nil)

		svc := settingsApp.New(repo, media, nil, &mockEventPublisher{}, nil)
		_, err := svc.SetPracticePageVideo(ctx, &domain.SetPracticePageVideoRequest{VideoURL: ""})

		require.NoError(t, err)
		media.AssertNotCalled(t, "DeleteObject", mock.Anything, mock.Anything)
	})
}

func TestGetMedia_ServesOnlyPracticePageVideos(t *testing.T) {
	ctx := context.Background()
	media := &mockSettingsMedia{}
	svc := settingsApp.New(&mockSettingsRepository{}, media, nil, &mockEventPublisher{}, nil)

	for _, key := range []string{
		"logo/company.png",
		"practice-page/video/../../logo/company.png",
		"practice-page/video/not-a-uuid.mp4",
		"practice-page/video/0b8a4a5e-6f0e-4f7b-9d55-6a9a0f3c1e11.mov",
		"practice-page/video/0b8a4a5e-6f0e-4f7b-9d55-6a9a0f3c1e11.mp4.png",
		"",
	} {
		_, err := svc.GetMedia(ctx, key, "")
		assert.ErrorIs(t, err, errs.ErrDomainNotFound, "key %q", key)
	}
	media.AssertNotCalled(t, "GetObject", mock.Anything, mock.Anything, mock.Anything)
}
