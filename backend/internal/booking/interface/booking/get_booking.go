package bookingHandler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	"github.com/google/uuid"
)

func (h *handler) GetBooking(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bookingID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr("invalid booking ID format"), http.StatusBadRequest)
		return
	}

	// Call service to get booking
	booking, err := h.svc.GetBooking(ctx, bookingID)
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

	// Convert to response DTO
	var cancellationReason *string
	if booking.CancellationReason != "" {
		cancellationReason = &booking.CancellationReason
	}

	// Derive plain-language-ready cancellation deadline and refund status
	// (issue 011). Best-effort: a failure here shouldn't hide the booking
	// itself, so log and continue with an empty financial status.
	financialStatus, err := h.svc.GetBookingFinancialStatus(ctx, booking)
	if err != nil {
		slog.WarnContext(ctx, "failed to derive booking financial status",
			"booking_id", booking.ID,
			"err", err,
		)
		financialStatus = &domain.BookingFinancialStatus{}
	}

	response := domain.BookingResponse{
		ID:                 booking.ID,
		AvailabilityID:     booking.AvailabilityID,
		ClientID:           booking.ClientID,
		PartnerID:          booking.PartnerID,
		RoomID:             booking.RoomID,
		ProductID:          booking.ProductID,
		SlotStartTime:      booking.SlotStartTime,
		SlotEndTime:        booking.SlotEndTime,
		Status:             booking.Status,
		TotalPriceCents:    booking.TotalPriceCents,
		Currency:           booking.Currency,
		PaymentStatus:      booking.PaymentStatus,
		PaymentIntentID:    booking.PaymentIntentID,
		ClientNotes:        booking.ClientNotes,
		PartnerNotes:       booking.PartnerNotes,
		CancellationReason: cancellationReason,
		CancelledAt:        booking.CancelledAt,
		CompletedAt:        booking.CompletedAt,
		GuestFirstName:     booking.GuestFirstName,
		GuestLastName:      booking.GuestLastName,
		GuestEmail:         booking.GuestEmail,
		GuestPhone:         booking.GuestPhone,
		Token:              booking.Token,

		BookingFinancialStatus: financialStatus,
	}

	// Surface the Stripe-hosted receipt URL once payment has succeeded.
	// Best-effort: a Stripe lookup failure should not fail the booking fetch.
	if booking.PaymentStatus == domain.PaymentStatusPaid && booking.PaymentIntentID != nil {
		paymentInfo, err := h.paymentService.RetrievePaymentIntent(ctx, *booking.PaymentIntentID)
		if err != nil {
			slog.WarnContext(ctx, "failed to retrieve payment intent for receipt url",
				"booking_id", booking.ID, "payment_intent_id", *booking.PaymentIntentID, "err", err)
		} else {
			response.ReceiptURL = paymentInfo.ReceiptURL
		}
	}

	httpx.RespondWithJSON(w, response, http.StatusOK)
}