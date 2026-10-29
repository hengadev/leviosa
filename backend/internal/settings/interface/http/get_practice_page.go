package http

import (
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
)

func (h *handler) GetPracticePagePartner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	logger.InfoContext(ctx, "Handler: Processing get practice page partner request",
		"operation", "get_practice_page_partner",
		"method", r.Method,
		"path", r.URL.Path)

	response, err := h.svc.GetPracticePagePartner(ctx)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "get practice page partner")
		return
	}

	logger.InfoContext(ctx, "Handler: Get practice page partner completed",
		"operation", "get_practice_page_partner",
		"status_code", http.StatusOK)

	httpx.RespondWithJSON(w, response, http.StatusOK)
}

func (h *handler) GetPracticePageProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	logger.InfoContext(ctx, "Handler: Processing get practice page product request",
		"operation", "get_practice_page_product",
		"method", r.Method,
		"path", r.URL.Path)

	response, err := h.svc.GetPracticePageProduct(ctx)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "get practice page product")
		return
	}

	logger.InfoContext(ctx, "Handler: Get practice page product completed",
		"operation", "get_practice_page_product",
		"status_code", http.StatusOK)

	httpx.RespondWithJSON(w, response, http.StatusOK)
}

func (h *handler) GetPracticePageVideo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	logger.InfoContext(ctx, "Handler: Processing get practice page video request",
		"operation", "get_practice_page_video",
		"method", r.Method,
		"path", r.URL.Path)

	response, err := h.svc.GetPracticePageVideo(ctx)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "get practice page video")
		return
	}

	logger.InfoContext(ctx, "Handler: Get practice page video completed",
		"operation", "get_practice_page_video",
		"status_code", http.StatusOK)

	httpx.RespondWithJSON(w, response, http.StatusOK)
}
