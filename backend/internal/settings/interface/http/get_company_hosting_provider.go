package http

import (
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
)

func (h *handler) GetCompanyHostingProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	logger.InfoContext(ctx, "Handler: Processing get company hosting provider request",
		"operation", "get_company_hosting_provider",
		"method", r.Method,
		"path", r.URL.Path)

	response, err := h.svc.GetCompanyHostingProvider(ctx)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "get company hosting provider")
		return
	}

	logger.InfoContext(ctx, "Handler: Get company hosting provider completed",
		"operation", "get_company_hosting_provider",
		"status_code", http.StatusOK)

	httpx.RespondWithJSON(w, response, http.StatusOK)
}
