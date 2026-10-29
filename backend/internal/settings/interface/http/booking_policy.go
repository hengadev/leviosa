package http

import (
	"context"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
)

func (h *handler) GetBookingPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	logger.InfoContext(ctx, "Handler: Processing get booking policy request",
		"operation", "get_booking_policy",
		"method", r.Method,
		"path", r.URL.Path)

	response, err := h.svc.GetBookingPolicy(ctx)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "get booking policy")
		return
	}

	httpx.RespondWithJSON(w, response, http.StatusOK)
}

func (h *handler) SetBookingPolicy(w http.ResponseWriter, r *http.Request) {
	// Same decoding, logging and error mapping as the practice page setters.
	setPracticePageSetting(w, r, domain.SetBookingPolicyRequest{},
		"set_booking_policy", "set booking policy",
		func(ctx context.Context, request *domain.SetBookingPolicyRequest) (*domain.SetBookingPolicyResponse, error) {
			return h.svc.SetBookingPolicy(ctx, request)
		})
}
