package bookingHandler

import (
	"errors"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/auth/session"
	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	"github.com/google/uuid"
)

// GetPartnerNextWorkingDay answers the partner's next working day and that
// day's appointment bookings (issue 037), for Aujourd'hui. Same
// authorization as the partner bookings list: the practitioner themselves or
// an administrator.
func (h *handler) GetPartnerNextWorkingDay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	partnerID, err := uuid.Parse(r.PathValue("partnerId"))
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("invalid partner ID format"), http.StatusBadRequest)
		return
	}

	sessionInfo, ok := session.SessionInfoFromContext(ctx)
	if !ok {
		httpx.RespondWithError(w, errs.ErrUnauthorized, http.StatusUnauthorized)
		return
	}
	if sessionInfo.UserID != partnerID && !sessionInfo.Role.IsAtLeast(identity.Administrator) {
		httpx.RespondWithError(w, errs.ErrForbidden, http.StatusForbidden)
		return
	}

	result, err := h.svc.GetPartnerNextWorkingDay(ctx, partnerID)
	if err != nil {
		var statusCode int
		switch {
		case errors.Is(err, errs.ErrRepositoryNotFound):
			statusCode = http.StatusNotFound
		default:
			statusCode = http.StatusInternalServerError
		}
		httpx.RespondWithError(w, err, statusCode)
		return
	}

	httpx.RespondWithJSON(w, result, http.StatusOK)
}
