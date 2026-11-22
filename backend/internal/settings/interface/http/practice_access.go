package http

import (
	"context"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
)

// Practice access handlers (admin-only, see routes.go). Neither the door code
// nor the instructions is logged: only the operation, method and path are.

func (h *handler) GetPracticeAccess(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	logger.InfoContext(ctx, "Handler: Processing get practice access request",
		"operation", "get_practice_access",
		"method", r.Method,
		"path", r.URL.Path)

	response, err := h.svc.GetPracticeAccess(ctx)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "get practice access")
		return
	}

	httpx.RespondWithJSON(w, response, http.StatusOK)
}

func (h *handler) SetPracticeAccess(w http.ResponseWriter, r *http.Request) {
	// Same decoding, logging and error mapping as the practice page setters.
	setPracticePageSetting(w, r, domain.SetPracticeAccessRequest{},
		"set_practice_access", "set practice access",
		func(ctx context.Context, request *domain.SetPracticeAccessRequest) (*domain.SetPracticeAccessResponse, error) {
			return h.svc.SetPracticeAccess(ctx, request)
		})
}
