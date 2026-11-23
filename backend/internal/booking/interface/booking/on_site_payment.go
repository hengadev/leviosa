package bookingHandler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/auth/session"
	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	"github.com/google/uuid"
)

// onSitePaymentConflicts are the refusals of the on-site payment rules,
// answered 409 with the domain's own message (the frontend maps it to
// French).
var onSitePaymentConflicts = []error{
	domain.ErrOnSitePaymentNotAllowed,
	domain.ErrOnSitePaymentBeforeStart,
	domain.ErrNothingDueOnSite,
	domain.ErrOnSitePaymentAlreadyRecorded,
	domain.ErrOnSitePaymentNotRecorded,
	domain.ErrOnSitePaymentUndoSameDayOnly,
}

// RecordOnSitePayment handles POST /bookings/{id}/on-site-payment: the
// client paid the amount due at the practice, on the card terminal.
func (h *handler) RecordOnSitePayment(w http.ResponseWriter, r *http.Request) {
	h.changeOnSitePayment(w, r, h.svc.RecordOnSitePayment)
}

// UndoOnSitePayment handles DELETE /bookings/{id}/on-site-payment: removes
// the recorded on-site payment, only on the Paris day it was recorded.
func (h *handler) UndoOnSitePayment(w http.ResponseWriter, r *http.Request) {
	h.changeOnSitePayment(w, r, h.svc.UndoOnSitePayment)
}

func (h *handler) changeOnSitePayment(w http.ResponseWriter, r *http.Request, change func(context.Context, uuid.UUID) (*domain.Booking, error)) {
	ctx := r.Context()

	bookingID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("invalid booking ID format"), http.StatusBadRequest)
		return
	}

	// Only the booking's practitioner, or an administrator, may record or
	// undo it; another practitioner's booking answers as unknown.
	sessionInfo, ok := session.SessionInfoFromContext(ctx)
	if !ok {
		httpx.RespondWithError(w, errs.ErrUnauthorized, http.StatusUnauthorized)
		return
	}
	existing, err := h.svc.GetBooking(ctx, bookingID)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			httpx.RespondWithError(w, errs.ErrRepositoryNotFound, http.StatusNotFound)
			return
		}
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}
	if sessionInfo.UserID != existing.PartnerID && !sessionInfo.Role.IsAtLeast(identity.Administrator) {
		httpx.RespondWithError(w, errs.ErrRepositoryNotFound, http.StatusNotFound)
		return
	}

	booking, err := change(ctx, bookingID)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			httpx.RespondWithError(w, errs.ErrRepositoryNotFound, http.StatusNotFound)
			return
		}
		for _, conflict := range onSitePaymentConflicts {
			if errors.Is(err, conflict) {
				httpx.RespondWithError(w, conflict, http.StatusConflict)
				return
			}
		}
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	response := cancelBookingResponse(booking)
	response.NoShowAt = booking.NoShowAt
	response.AmountDueOnSiteCents = booking.AmountDueOnSiteCents()
	response.AmountPaidOnSiteCents = booking.AmountPaidOnSiteCents
	response.PaidOnSiteAt = booking.PaidOnSiteAt
	httpx.RespondWithJSON(w, response, http.StatusOK)
}
