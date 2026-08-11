package http

import (
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
)

func (h *handler) GetCompanyRegistrationNumber(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	logger.InfoContext(ctx, "Handler: Processing get company registration number request",
		"operation", "get_company_registration_number",
		"method", r.Method,
		"path", r.URL.Path)

	response, err := h.svc.GetCompanyRegistrationNumber(ctx)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "get company registration number")
		return
	}

	logger.InfoContext(ctx, "Handler: Get company registration number completed",
		"operation", "get_company_registration_number",
		"status_code", http.StatusOK)

	httpx.RespondWithJSON(w, response, http.StatusOK)
}
