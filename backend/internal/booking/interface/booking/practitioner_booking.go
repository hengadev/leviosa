package bookingHandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/auth/session"
	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	"github.com/Leviosa-care/leviosa/backend/internal/common/validation"

	"github.com/google/uuid"
)

// Practitioner booking (issue 041, ADR-0021): the practitioner books a
// client who called, from the "Nouvelle réservation" form.

// practitionerBookingResponse is the booking just made by the practitioner.
type practitionerBookingResponse struct {
	ID                   uuid.UUID            `json:"id"`
	AvailabilityID       uuid.UUID            `json:"availability_id"`
	PartnerID            uuid.UUID            `json:"partner_id"`
	ProductID            uuid.UUID            `json:"product_id"`
	SlotStartTime        time.Time            `json:"slot_start_time"`
	SlotEndTime          time.Time            `json:"slot_end_time"`
	Status               domain.BookingStatus `json:"status"`
	Origin               domain.BookingOrigin `json:"origin"`
	TotalPriceCents      int                  `json:"total_price_cents"`
	Currency             string               `json:"currency"`
	AmountDueOnSiteCents int                  `json:"amount_due_on_site_cents"`
}

// CreatePractitionerBooking handles POST /partners/bookings: the caller
// books a slot of their own agenda (or, as an administrator, of the
// practitioner's) for a client who called. 201 with the confirmed booking;
// 400 on an invalid body, client or slot; 403 on another partner's
// availability; 409 when the slot was taken meanwhile.
func (h *handler) CreatePractitionerBooking(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sessionInfo, ok := session.SessionInfoFromContext(ctx)
	if !ok || sessionInfo == nil {
		httpx.RespondWithError(w, errs.ErrUnauthorized, http.StatusUnauthorized)
		return
	}

	var input domain.PractitionerBookingInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr(fmt.Sprintf("invalid request body: %v", err)), http.StatusBadRequest)
		return
	}
	input.Normalize()
	if input.GuestPhone != "" {
		if err := validation.ValidatePhone(input.GuestPhone); err != nil {
			httpx.RespondWithError(w, errs.NewInvalidValueErr(fmt.Sprintf("guest_phone: %v", err)), http.StatusBadRequest)
			return
		}
	}
	if input.GuestEmail != "" {
		if err := validation.ValidateEmail(input.GuestEmail); err != nil {
			httpx.RespondWithError(w, errs.NewInvalidValueErr(fmt.Sprintf("guest_email: %v", err)), http.StatusBadRequest)
			return
		}
	}

	booking, err := h.svc.CreatePractitionerBooking(ctx, input, sessionInfo.UserID, sessionInfo.Role.IsAtLeast(identity.Administrator))
	if err != nil {
		httpx.RespondWithError(w, err, practitionerBookingStatus(err))
		return
	}

	httpx.RespondWithJSON(w, practitionerBookingResponse{
		ID:                   booking.ID,
		AvailabilityID:       booking.AvailabilityID,
		PartnerID:            booking.PartnerID,
		ProductID:            booking.ProductID,
		SlotStartTime:        booking.SlotStartTime,
		SlotEndTime:          booking.SlotEndTime,
		Status:               booking.Status,
		Origin:               booking.Origin,
		TotalPriceCents:      booking.TotalPriceCents,
		Currency:             booking.Currency,
		AmountDueOnSiteCents: booking.AmountDueOnSiteCents(),
	}, http.StatusCreated)
}

// GetPractitionerProducts handles GET /partners/booking-products/{partnerId}:
// the products the practitioner can book (partnerId is the partner's user
// ID). The partner themselves or an administrator only.
func (h *handler) GetPractitionerProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	partnerID, ok := h.authorizedPartner(w, r)
	if !ok {
		return
	}
	products, err := h.svc.GetPractitionerProducts(ctx, partnerID)
	if err != nil {
		httpx.RespondWithError(w, err, practitionerBookingStatus(err))
		return
	}
	httpx.RespondWithJSON(w, products, http.StatusOK)
}

// GetPractitionerFreeSlots handles
//
//	GET /partners/booking-slots/{partnerId}?product_id=<uuid>&from=YYYY-MM-DD&to=YYYY-MM-DD
//
// the free slots of the practitioner's form: the public slot endpoint's
// answer without the minimum notice and the booking horizon (partnerId is
// the partner's user ID; from and to are Europe/Paris dates, inclusive).
// The partner themselves or an administrator only.
func (h *handler) GetPractitionerFreeSlots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	partnerID, ok := h.authorizedPartner(w, r)
	if !ok {
		return
	}
	query := r.URL.Query()
	productID, err := uuid.Parse(query.Get("product_id"))
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("product_id: must be a UUID"), http.StatusBadRequest)
		return
	}
	fromDate, err := time.ParseInLocation(time.DateOnly, query.Get("from"), domain.PracticeLocation)
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("from: must be a date (YYYY-MM-DD)"), http.StatusBadRequest)
		return
	}
	toDate, err := time.ParseInLocation(time.DateOnly, query.Get("to"), domain.PracticeLocation)
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("to: must be a date (YYYY-MM-DD)"), http.StatusBadRequest)
		return
	}
	if toDate.Before(fromDate) {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("to: must not be before from"), http.StatusBadRequest)
		return
	}
	toExclusive := toDate.AddDate(0, 0, 1)
	if toExclusive.After(fromDate.AddDate(0, 0, maxSlotRangeDays)) {
		httpx.RespondWithError(w, errs.NewInvalidValueErr(fmt.Sprintf("range: at most %d days", maxSlotRangeDays)), http.StatusBadRequest)
		return
	}

	slots, err := h.svc.GetPractitionerFreeSlots(ctx, partnerID, productID, fromDate, toExclusive)
	if err != nil {
		httpx.RespondWithError(w, err, practitionerBookingStatus(err))
		return
	}
	httpx.RespondWithJSON(w, slots.ToResponse(), http.StatusOK)
}

// authorizedPartner reads the {partnerId} path value (a partner's user ID)
// and checks the caller is that partner or an administrator, answering the
// error itself when not.
func (h *handler) authorizedPartner(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	partnerID, err := uuid.Parse(r.PathValue("partnerId"))
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("invalid partner ID format"), http.StatusBadRequest)
		return uuid.Nil, false
	}
	sessionInfo, ok := session.SessionInfoFromContext(r.Context())
	if !ok || sessionInfo == nil {
		httpx.RespondWithError(w, errs.ErrUnauthorized, http.StatusUnauthorized)
		return uuid.Nil, false
	}
	if sessionInfo.UserID != partnerID && !sessionInfo.Role.IsAtLeast(identity.Administrator) {
		httpx.RespondWithError(w, errs.ErrForbidden, http.StatusForbidden)
		return uuid.Nil, false
	}
	return partnerID, true
}

// practitionerBookingStatus maps the practitioner booking errors to HTTP.
func practitionerBookingStatus(err error) int {
	switch {
	case errors.Is(err, errs.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, errs.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, errs.ErrInvalidValue), errors.Is(err, errs.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, errs.ErrRepositoryNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
