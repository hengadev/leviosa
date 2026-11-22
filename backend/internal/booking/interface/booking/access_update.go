package bookingHandler

import (
	"errors"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
)

// GetAccessUpdateRecipients answers how many upcoming confirmed bookings an
// access update would reach, and how many have no client email.
func (h *handler) GetAccessUpdateRecipients(w http.ResponseWriter, r *http.Request) {
	counts, err := h.svc.GetAccessUpdateRecipients(r.Context())
	if err != nil {
		httpx.RespondWithError(w, err, accessUpdateStatus(err))
		return
	}
	httpx.RespondWithJSON(w, counts, http.StatusOK)
}

// SendAccessUpdate emails the current practice access to the clients of
// upcoming confirmed bookings and answers the counts. 400 when no door code
// is set.
func (h *handler) SendAccessUpdate(w http.ResponseWriter, r *http.Request) {
	counts, err := h.svc.SendAccessUpdate(r.Context())
	if err != nil {
		httpx.RespondWithError(w, err, accessUpdateStatus(err))
		return
	}
	httpx.RespondWithJSON(w, counts, http.StatusOK)
}

func accessUpdateStatus(err error) int {
	switch {
	case errors.Is(err, errs.ErrInvalidValue):
		return http.StatusBadRequest
	case errors.Is(err, errs.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, errs.ErrForbidden):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
