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

	// Refund amount/date: today's refund flow (RefundBooking) only ever
	// issues a full refund, so the amount is inferred from total_price_cents
	// once payment_status flips to refunded — the same inference the admin
	// financial summary already relies on. RefundedAt falls back to the
	// booking's UpdatedAt, which RefundBooking bumps at refund time. Both
	// fields are written generically (see BookingFinancialStatus) so a future
	// graduated refund policy (issue 017) can populate them from its own
	// stored amount/timestamp without any change on the reading side.
	if booking.PaymentStatus == domain.PaymentStatusRefunded {
		amount := booking.TotalPriceCents
		refundedAt := booking.UpdatedAt
		status.RefundAmountCents = &amount
		status.RefundedAt = &refundedAt
	}

	return status, nil
}
