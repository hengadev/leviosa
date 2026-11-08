package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/ports"

	"github.com/stretchr/testify/assert"
)

// mediaService answers GetMedia only; the embedded interface is nil.
type mediaService struct {
	ports.SettingsService
	gotKey, gotRange string
	object           *domain.MediaObject
	err              error
}

func (s *mediaService) GetMedia(_ context.Context, key, rangeHeader string) (*domain.MediaObject, error) {
	s.gotKey, s.gotRange = key, rangeHeader
	return s.object, s.err
}

func serveMedia(svc *mediaService, method, path, rangeHeader string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+GetMediaEndpoint, (&handler{svc: svc}).GetMedia)
	req := httptest.NewRequest(method, path, nil)
	if rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

const videoPath = "/media/practice-page/video/0b8a4a5e-6f0e-4f7b-9d55-6a9a0f3c1e11.mp4"

func TestGetMedia(t *testing.T) {
	t.Run("streams the whole video, cacheable for good", func(t *testing.T) {
		svc := &mediaService{object: &domain.MediaObject{
			Body: io.NopCloser(strings.NewReader("video")), ContentType: "video/mp4", ContentLength: 5, ETag: `"abc"`,
		}}
		rec := serveMedia(svc, http.MethodGet, videoPath, "")

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "practice-page/video/0b8a4a5e-6f0e-4f7b-9d55-6a9a0f3c1e11.mp4", svc.gotKey)
		assert.Equal(t, "video", rec.Body.String())
		assert.Equal(t, "video/mp4", rec.Header().Get("Content-Type"))
		assert.Equal(t, "5", rec.Header().Get("Content-Length"))
		assert.Equal(t, "bytes", rec.Header().Get("Accept-Ranges"))
		assert.Contains(t, rec.Header().Get("Cache-Control"), "immutable")
	})

	t.Run("answers a range with 206", func(t *testing.T) {
		svc := &mediaService{object: &domain.MediaObject{
			Body: io.NopCloser(strings.NewReader("de")), ContentType: "video/mp4", ContentLength: 2, ContentRange: "bytes 3-4/5",
		}}
		rec := serveMedia(svc, http.MethodGet, videoPath, "bytes=3-4")

		assert.Equal(t, http.StatusPartialContent, rec.Code)
		assert.Equal(t, "bytes=3-4", svc.gotRange)
		assert.Equal(t, "bytes 3-4/5", rec.Header().Get("Content-Range"))
		assert.Equal(t, "de", rec.Body.String())
	})

	t.Run("serves several ranges as the whole video", func(t *testing.T) {
		svc := &mediaService{object: &domain.MediaObject{Body: io.NopCloser(strings.NewReader("video")), ContentLength: 5}}
		rec := serveMedia(svc, http.MethodGet, videoPath, "bytes=0-1,3-4")

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, svc.gotRange)
	})

	t.Run("HEAD sends the headers only", func(t *testing.T) {
		svc := &mediaService{object: &domain.MediaObject{Body: io.NopCloser(strings.NewReader("video")), ContentLength: 5}}
		rec := serveMedia(svc, http.MethodHead, videoPath, "")

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "5", rec.Header().Get("Content-Length"))
		assert.Empty(t, rec.Body.String())
	})

	t.Run("an unknown key is 404", func(t *testing.T) {
		rec := serveMedia(&mediaService{err: errs.NewNotFoundErr(io.EOF, "media")}, http.MethodGet, "/media/logo/x.png", "")
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("a range outside the video is 416", func(t *testing.T) {
		rec := serveMedia(&mediaService{err: domain.ErrMediaRangeNotSatisfiable}, http.MethodGet, videoPath, "bytes=100-")
		assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, rec.Code)
	})
}
