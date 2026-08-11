package http

import (
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
)

func (h *handler) GetCompanyVatNumber(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	logger.InfoContext(ctx, "Handler: Processing get company VAT number request",
		"operation", "get_company_vat_number",
		"method", r.Method,
		"path", r.URL.Path)

	response, err := h.svc.GetCompanyVatNumber(ctx)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "get company VAT number")
		return
	}

	logger.InfoContext(ctx, "Handler: Get company VAT number completed",
		"operation", "get_company_vat_number",
		"status_code", http.StatusOK)

	httpx.RespondWithJSON(w, response, http.StatusOK)
}
