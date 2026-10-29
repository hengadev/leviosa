package booking

import (
	"context"
	"errors"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

// GetBookingFinancialStatus computes the fields a client UI needs to render
// plain-language payment/refund status and the cancellation deadline ahead of
// the cancel action (issue 011). It does not change any cancellation or
// refund policy — CancelBooking/RefundBooking remain the source of truth for
// enforcement; this method only derives display data from the same inputs.
func (s *BookingService) GetBookingFinancialStatus(ctx context.Context, booking *domain.Booking) (*domain.BookingFinancialStatus, error) {
	status := &domain.BookingFinancialStatus{}

	// A cancellation deadline is only meaningful while the booking is still
	// cancellable; once cancelled/completed/no-show there's nothing left to
	// warn about.
	if booking.IsCancellable() {
		product, err := s.productService.GetProductByID(ctx, booking.ProductID.String())
		if err != nil && !errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, err
		}
		if product != nil && product.CancellationHours > 0 {
			deadline := booking.SlotStartTime.Add(-time.Duration(product.CancellationHours) * time.Hour)
			hours := product.CancellationHours
			status.CancellationDeadline = &deadline
			status.CancellationHours = &hours
		}
	}

	// Refund amount/date: a cancellation records what it refunded (issue
	// 017); a manual full refund (RefundBooking) or a lost dispute does not,
	// so the amount is then inferred from total_price_cents once
	// payment_status is refunded. RefundedAt falls back to the booking's
	// UpdatedAt, bumped at refund time.
	switch {
	case booking.RefundedAmountCents != nil && *booking.RefundedAmountCents > 0:
		amount := *booking.RefundedAmountCents
		refundedAt := booking.UpdatedAt
		status.RefundAmountCents = &amount
		status.RefundedAt = &refundedAt
	case booking.PaymentStatus == domain.PaymentStatusRefunded:
		amount := booking.TotalPriceCents
		refundedAt := booking.UpdatedAt
		status.RefundAmountCents = &amount
		status.RefundedAt = &refundedAt
	}

	return status, nil
}
