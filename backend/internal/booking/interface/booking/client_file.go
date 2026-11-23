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

// GetBookingClientFile returns the client file of the booking's client
// (issue 042): contact details plus upcoming and past sessions with the
// booking's practitioner. Same authorization as the partner bookings list,
// but checked against the booking's partner: another practitioner's booking
// is a 404, not a 403.
func (h *handler) GetBookingClientFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bookingID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("invalid booking ID format"), http.StatusBadRequest)
		return
	}

	sessionInfo, ok := session.SessionInfoFromContext(ctx)
	if !ok {
		httpx.RespondWithError(w, errs.ErrUnauthorized, http.StatusUnauthorized)
		return
	}

	booking, err := h.svc.GetBooking(ctx, bookingID)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			httpx.RespondWithError(w, errs.ErrRepositoryNotFound, http.StatusNotFound)
			return
		}
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}
	if sessionInfo.UserID != booking.PartnerID && !sessionInfo.Role.IsAtLeast(identity.Administrator) {
		httpx.RespondWithError(w, errs.ErrRepositoryNotFound, http.StatusNotFound)
		return
	}

	file, err := h.svc.GetClientFile(ctx, booking.PartnerID, bookingID)
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

	httpx.RespondWithJSON(w, file, http.StatusOK)
}

// SearchPartnerClients returns the practitioner's distinct clients matching
// a name or phone fragment (issue 043 wires it into the booking form).
// Same authorization as the partner bookings list.
func (h *handler) SearchPartnerClients(w http.ResponseWriter, r *http.Request) {
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

	clients, err := h.svc.SearchPartnerClients(ctx, partnerID, r.URL.Query().Get("q"))
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	httpx.RespondWithJSON(w, clients, http.StatusOK)
}
