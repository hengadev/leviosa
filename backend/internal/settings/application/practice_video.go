package settings

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
)

// UploadPracticePageVideo stores an uploaded video in the media bucket and
// makes it the practice page video. The format is read from the file's first
// bytes, never from what the browser declared. The previous upload, if any, is
// deleted once the setting points at the new one.
func (s *SettingsService) UploadPracticePageVideo(ctx context.Context, file io.ReadSeeker, size int64) (*domain.UploadPracticePageVideoResponse, error) {
	if file == nil || size <= 0 {
		return nil, errs.NewInvalidValueErr("video file is required")
	}
	if size > domain.PracticePageVideoMaxBytes {
		return nil, errs.NewInvalidValueErr(fmt.Sprintf("video file exceeds %d MB", domain.PracticePageVideoMaxBytes>>20))
	}

	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, errs.NewInvalidValueErr("video file could not be read")
	}
	contentType := http.DetectContentType(head[:n])
	ext, ok := domain.PracticePageVideoExtension(contentType)
	if !ok {
		return nil, errs.NewInvalidValueErr("video must be an MP4 or WebM file")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rewind video file: %w", err)
	}

	key := domain.NewPracticePageVideoKey(ext)
	if err := s.media.PutObject(ctx, key, file, size, contentType); err != nil {
		return nil, errs.NewExternalServiceErr(err, "upload practice page video")
	}

	url := domain.PracticePageVideoURL(key)
	if err := s.replacePracticePageVideo(ctx, url); err != nil {
		// The setting still points at the previous video: drop the orphan.
		s.deleteMedia(ctx, key)
		return nil, err
	}
	return &domain.UploadPracticePageVideoResponse{VideoURL: url}, nil
}

// GetMedia reads a media object, or the byte range rangeHeader asks for, for
// the public media route. Only uploaded practice page videos are served: any
// other key is not found, whatever the bucket holds.
func (s *SettingsService) GetMedia(ctx context.Context, key, rangeHeader string) (*domain.MediaObject, error) {
	if !domain.IsPracticePageVideoKey(key) {
		return nil, errs.NewNotFoundErr(errors.New("unknown media"), "media")
	}
	return s.media.GetObject(ctx, key, rangeHeader)
}

// replacePracticePageVideo sets the practice page video and deletes the video
// it replaces when that one was uploaded here.
func (s *SettingsService) replacePracticePageVideo(ctx context.Context, value string) error {
	var previous string
	if current, err := s.repo.GetString(ctx, settings.PracticePageVideoURL); err == nil {
		previous = current.Value
	}

	setting := NewSetting(settings.PracticePageVideoURL, value)
	if err := s.repo.SetString(ctx, setting); err != nil {
		return fmt.Errorf("set practice page video: %w", err)
	}
	s.auditFieldUpdate(ctx, settings.PracticePageVideoURL, setting.Value)

	if key, ok := domain.PracticePageVideoKeyFromURL(previous); ok && previous != value {
		s.deleteMedia(ctx, key)
	}
	return nil
}

// deleteMedia removes an object that is no longer used. A failure only leaves
// an unreferenced object behind, so it is logged, not returned.
func (s *SettingsService) deleteMedia(ctx context.Context, key string) {
	if s.media == nil {
		return
	}
	if err := s.media.DeleteObject(ctx, key); err != nil {
		logger, lerr := ctxutil.GetLoggerFromContext(ctx)
		if lerr != nil {
			logger = slog.Default()
		}
		logger.WarnContext(ctx, "Service: could not delete a replaced media object", "key", key, "error", err)
	}
}
