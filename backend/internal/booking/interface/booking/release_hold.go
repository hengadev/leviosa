package bookingHandler

import (
	"errors"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
)

// ReleaseHold frees the slot of a guest booking still awaiting payment when
// the client leaves the payment step (ADR-0018). Unauthenticated: the booking
// token (?token=xxx) proves the caller made the booking. 204 when nothing is
// held any more (released now, or confirmed/expired already); 409 when a
// payment is under way and the hold is kept.
func (h *handler) ReleaseHold(w http.ResponseWriter, r *http.Request) {
	token := httpx.FormString(r, "token")
	if token == "" {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("missing booking token"), http.StatusBadRequest)
		return
	}

	if err := h.svc.ReleaseHoldByToken(r.Context(), token); err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, errs.ErrUnauthorized):
			status = http.StatusUnauthorized
		case errors.Is(err, errs.ErrConflict):
			status = http.StatusConflict
		}
		httpx.RespondWithError(w, err, status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
