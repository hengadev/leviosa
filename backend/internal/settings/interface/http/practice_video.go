package http

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
)

// UploadPracticePageVideo receives the practice page video as the "video"
// field of a multipart form. The form is spooled to a temporary file past a
// few megabytes, so a 50 MB upload does not sit in memory.
func (h *handler) UploadPracticePageVideo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	logger.InfoContext(ctx, "Handler: Processing upload practice page video request",
		"operation", "upload_practice_page_video",
		"method", r.Method,
		"path", r.URL.Path)

	// The server's 15 s read timeout covers the whole body: far too short for
	// 50 MB on a home connection. This request alone gets longer.
	if err := http.NewResponseController(w).SetReadDeadline(time.Now().Add(10 * time.Minute)); err != nil {
		logger.WarnContext(ctx, "Handler: could not extend the upload's read deadline", "error", err)
	}

	// The file plus room for the form's boundaries and headers.
	r.Body = http.MaxBytesReader(w, r.Body, domain.PracticePageVideoMaxBytes+1<<20)
	const maxFormMemory = 8 << 20
	if err := r.ParseMultipartForm(maxFormMemory); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			httpx.RespondWithError(w, fmt.Errorf("video file exceeds %d MB", domain.PracticePageVideoMaxBytes>>20), http.StatusRequestEntityTooLarge)
			return
		}
		httpx.RespondWithError(w, errors.New("failed to parse form data"), http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("video")
	if err != nil {
		httpx.RespondWithError(w, errors.New("required form file 'video' is missing"), http.StatusBadRequest)
		return
	}
	defer file.Close()

	response, err := h.svc.UploadPracticePageVideo(ctx, file, header.Size)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "upload practice page video")
		return
	}

	logger.InfoContext(ctx, "Handler: Upload practice page video completed",
		"operation", "upload_practice_page_video",
		"size_bytes", header.Size,
		"status_code", http.StatusCreated)

	httpx.RespondWithJSON(w, response, http.StatusCreated)
}

// GetMedia streams a media object to visitors, with byte ranges so browsers
// can seek and start playing before the whole video has arrived. Each upload
// has its own key, so the response can be cached for good.
func (h *handler) GetMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := r.PathValue("key")

	// S3 serves one range per request; anything else gets the whole object,
	// which is a valid answer to any Range header.
	rangeHeader := r.Header.Get("Range")
	if !strings.HasPrefix(rangeHeader, "bytes=") || strings.Contains(rangeHeader, ",") {
		rangeHeader = ""
	}

	object, err := h.svc.GetMedia(ctx, key, rangeHeader)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMediaRangeNotSatisfiable):
			w.Header().Set("Accept-Ranges", "bytes")
			http.Error(w, "range not satisfiable", http.StatusRequestedRangeNotSatisfiable)
		case errors.Is(err, errs.ErrDomainNotFound), errors.Is(err, errs.ErrRepositoryNotFound):
			http.NotFound(w, r)
		default:
			if logger, lerr := ctxutil.GetLoggerFromContext(ctx); lerr == nil {
				logger.ErrorContext(ctx, "Handler: get media failed", "key", key, "error", err)
			}
			http.Error(w, "media unavailable", http.StatusBadGateway)
		}
		return
	}
	defer object.Body.Close()

	header := w.Header()
	header.Set("Accept-Ranges", "bytes")
	header.Set("Cache-Control", "public, max-age=31536000, immutable")
	header.Set("X-Content-Type-Options", "nosniff")
	if object.ContentType != "" {
		header.Set("Content-Type", object.ContentType)
	}
	if object.ETag != "" {
		header.Set("ETag", object.ETag)
		if r.Header.Get("If-None-Match") == object.ETag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	if object.LastModified != nil {
		header.Set("Last-Modified", object.LastModified.UTC().Format(http.TimeFormat))
	}
	header.Set("Content-Length", fmt.Sprint(object.ContentLength))

	status := http.StatusOK
	if object.ContentRange != "" {
		header.Set("Content-Range", object.ContentRange)
		status = http.StatusPartialContent
	}
	w.WriteHeader(status)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = io.Copy(w, object.Body)
}
