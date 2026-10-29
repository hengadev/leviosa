package booking

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	catalogDomain "github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/audit"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
)

// Client cancellation with a tiered refund (ADR-0014, issue 017). A client
// can always cancel a confirmed booking; the product's cancellation tiers,
// applied to the notice given, decide how much of the amount paid online is
// refunded. The refund preview on the booking page and the cancellation use
// the same rule (cancellationRefund) and the same clock (s.now).

// CancelBookingByToken cancels a booking as the client, from the booking
// link: it verifies the signed booking token, ensures it matches the
// requested booking ID, and returns the cancelled booking as the booking page
// shows it.
func (s *BookingService) CancelBookingByToken(ctx context.Context, bookingID uuid.UUID, token string) (*domain.PublicBookingLookupResponse, error) {
	if len(s.tokenSecret) == 0 {
		return nil, fmt.Errorf("booking token feature is not configured")
	}

	tokenBookingID, err := domain.VerifyBookingToken(token, s.tokenSecret)
	if err != nil {
		if domain.IsBookingTokenError(err) {
			return nil, errs.NewUnauthorizedErr(err.Error())
		}
		return nil, fmt.Errorf("verify booking token for cancel: %w", err)
	}

	// The token must be scoped to the exact booking being cancelled.
	if tokenBookingID != bookingID {
		return nil, errs.NewUnauthorizedErr("token does not match the requested booking")
	}

	booking, product, err := s.cancelByClient(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	return s.publicResponse(ctx, booking, product), nil
}

// CancelBooking cancels a booking as the client, from their account. Same
// rule as the booking link.
func (s *BookingService) CancelBooking(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	booking, _, err := s.cancelByClient(ctx, id)
	return booking, err
}

// errBookingNotCancellable refuses cancelling a booking that is not
// confirmed: awaiting payment, expired, completed, no-show, or already
// cancelled (a second cancel request refunds nothing more).
func errBookingNotCancellable(status domain.BookingStatus) error {
	return errs.NewInvalidValueErr("booking cannot be cancelled: current status is " + string(status))
}

// cancelByClient is the client cancellation: the refund the product's tiers
// give on the notice (cancelAs).
func (s *BookingService) cancelByClient(ctx context.Context, id uuid.UUID) (*domain.Booking, *catalogDomain.ProductRes, error) {
	booking, product, err := s.loadCancellable(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return s.cancelAs(ctx, booking, product, domain.CancelledByClient)
}

// CancelBookingByPractitioner cancels a booking as a practitioner
// cancellation (CONTEXT.md), from the partner agenda: whatever the product's
// tiers, the client gets back everything they paid online, and is emailed.
// Only the booking's partner may do it, or an administrator
// (callerIsAdmin); anyone else gets ErrForbidden. Only a confirmed booking
// can be cancelled, so a second request refunds nothing more.
func (s *BookingService) CancelBookingByPractitioner(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerIsAdmin bool) (*domain.Booking, error) {
	booking, err := s.loadBooking(ctx, id)
	if err != nil {
		return nil, err
	}
	if booking.PartnerID != callerID && !callerIsAdmin {
		return nil, errs.NewForbiddenErr("booking belongs to another partner")
	}
	if !booking.IsCancellable() {
		return nil, errBookingNotCancellable(booking.Status)
	}
	product, err := s.bookingProduct(ctx, booking)
	if err != nil {
		return nil, err
	}
	booking, _, err = s.cancelAs(ctx, booking, product, domain.CancelledByPractitioner)
	return booking, err
}

// loadCancellable reads a booking that can be cancelled, with its product.
func (s *BookingService) loadCancellable(ctx context.Context, id uuid.UUID) (*domain.Booking, *catalogDomain.ProductRes, error) {
	booking, err := s.loadBooking(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if !booking.IsCancellable() {
		return nil, nil, errBookingNotCancellable(booking.Status)
	}
	product, err := s.bookingProduct(ctx, booking)
	if err != nil {
		return nil, nil, err
	}
	return booking, product, nil
}

// cancelAs cancels a confirmed booking on behalf of actor: the booking is
// claimed as cancelled first (only one of two concurrent requests gets it),
// then the refund the actor's rule gives (cancellationRefund: the tiers for
// the client, everything for the practitioner) is made through the payment
// port (none when it is zero). A failed refund releases the claim and fails
// the request, so the cancellation can be tried again and the booking never
// ends up cancelled without its refund. The slot is offered again and the
// cancellation emails are sent.
func (s *BookingService) cancelAs(ctx context.Context, booking *domain.Booking, product *catalogDomain.ProductRes, actor domain.CancellationActor) (*domain.Booking, *catalogDomain.ProductRes, error) {
	now := s.now()
	refund := s.cancellationRefund(booking, product, actor, now)

	claimed, err := s.bookingRepo.ClaimCancellation(ctx, booking.ID, actor, now, refund.AmountCents)
	if err != nil {
		return nil, nil, fmt.Errorf("claim booking cancellation: %w", err)
	}
	if !claimed {
		// Cancelled (or otherwise moved) by a concurrent request.
		current, err := s.loadBooking(ctx, booking.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, errBookingNotCancellable(current.Status)
	}

	var refundID string
	if refund.AmountCents > 0 {
		refundID, err = s.paymentService.RefundPayment(ctx, *booking.PaymentIntentID, refund.AmountCents, refundReason(actor))
		if err != nil {
			if releaseErr := s.bookingRepo.ReleaseCancellation(ctx, booking.ID); releaseErr != nil {
				slog.ErrorContext(ctx, "failed to release the cancellation claim after a failed refund",
					"booking_id", booking.ID, "error", releaseErr)
			}
			return nil, nil, fmt.Errorf("refund cancelled booking: %w", err)
		}

		paymentStatus := domain.PaymentStatusPartiallyRefunded
		if refund.AmountCents >= amountPaidOnline(booking, product) {
			paymentStatus = domain.PaymentStatusRefunded
		}
		if err := s.bookingRepo.MarkCancellationRefunded(ctx, booking.ID, paymentStatus); err != nil {
			// The money is back and the refunded amount is recorded; only
			// the payment status lags. Logged for follow-up.
			slog.ErrorContext(ctx, "failed to record the payment status of a cancellation refund",
				"booking_id", booking.ID, "refund_id", refundID, "error", err)
		}
		booking.PaymentStatus = paymentStatus
	}

	booking.Status = domain.BookingStatusCancelled
	booking.CancelledAt = &now
	booking.CancelledBy = actor
	refunded := refund.AmountCents
	booking.RefundedAmountCents = &refunded

	actorID, actorRole := audit.ActorFromContext(ctx)
	if actorID == uuid.Nil {
		actorRole = "guest" // from the booking link, no session
	}
	audit.WriteOrLog(ctx, s.audit, audit.Entry{
		ActorID:    actorID,
		ActorRole:  actorRole,
		Action:     "booking.cancelled",
		TargetType: "booking",
		TargetID:   booking.ID.String(),
		Payload: map[string]any{
			"cancelled_by":      string(actor),
			"refund_id":         refundID,
			"refund_cents":      refund.AmountCents,
			"refund_percentage": refund.Percentage,
		},
	})

	s.releaseAvailability(ctx, booking)
	s.sendCancellationNotice(ctx, booking, product)
	return booking, product, nil
}

// refundReason is the reason given to the payment provider for a
// cancellation refund. Stripe has no reason for a merchant-side
// cancellation; requested_by_customer is kept for the client's own.
func refundReason(actor domain.CancellationActor) string {
	if actor == domain.CancelledByClient {
		return ports.RefundReasonRequestedByCustomer
	}
	return ""
}

// AdminOverrideCancelBooking lets an admin cancel a booking with a reason,
// without any automatic refund (the admin refunds manually when needed,
// RefundBooking). The route is admin-gated (RequireAdmin), and every
// override writes an audit entry.
func (s *BookingService) AdminOverrideCancelBooking(ctx context.Context, id uuid.UUID, reason string) (*domain.Booking, error) {
	booking, err := s.loadBooking(ctx, id)
	if err != nil {
		return nil, err
	}
	if !booking.IsCancellable() {
		return nil, errBookingNotCancellable(booking.Status)
	}

	product, err := s.bookingProduct(ctx, booking)
	if err != nil {
		return nil, err
	}

	if err := booking.Cancel(reason); err != nil {
		return nil, fmt.Errorf("cancel booking: %w", err)
	}

	bookingEncx, err := domain.ProcessBookingEncx(ctx, s.crypto, booking)
	if err != nil {
		return nil, fmt.Errorf("encrypt booking: %w", err)
	}
	if err := s.bookingRepo.Update(ctx, bookingEncx); err != nil {
		return nil, fmt.Errorf("update cancelled booking: %w", err)
	}

	actorID, actorRole := audit.ActorFromContext(ctx)
	audit.WriteOrLog(ctx, s.audit, audit.Entry{
		ActorID:    actorID,
		ActorRole:  actorRole,
		Action:     "booking.admin_override",
		TargetType: "booking",
		TargetID:   booking.ID.String(),
		Payload: map[string]any{
			"reason":                       reason,
			"bypassed_cancellation_policy": true,
		},
	})

	s.releaseAvailability(ctx, booking)
	s.sendCancellationNotice(ctx, booking, product)
	return booking, nil
}

// cancellationRefund is the refund a cancellation by actor at now gives on
// the booking: the product's tiers applied to the amount paid online. Only a
// paid booking with a PaymentIntent has anything to refund.
func (s *BookingService) cancellationRefund(booking *domain.Booking, product *catalogDomain.ProductRes, actor domain.CancellationActor, now time.Time) domain.CancellationRefund {
	paid := 0
	if booking.PaymentStatus == domain.PaymentStatusPaid && booking.PaymentIntentID != nil {
		paid = amountPaidOnline(booking, product)
	}
	var tiers []domain.RefundTier
	if product != nil {
		for _, tier := range product.CancellationTiers {
			tiers = append(tiers, domain.RefundTier{MinHoursBefore: tier.MinHoursBefore, RefundPercentage: tier.RefundPercentage})
		}
	}
	return domain.ComputeCancellationRefund(actor, tiers, paid, booking.SlotStartTime, now)
}

// amountPaidOnline is what the client paid online for the booking: as
// recorded when it was confirmed, or, for bookings confirmed before that was
// recorded, the product's current deposit applied to the booking's total
// price (the full price without a deposit, or without the product). The
// fallback reads no payment provider, so the preview stays cheap and equal
// to the refund; it can differ from the charge only if the deposit changed
// since the booking was made.
func amountPaidOnline(booking *domain.Booking, product *catalogDomain.ProductRes) int {
	if booking.AmountPaidOnlineCents != nil {
		return *booking.AmountPaidOnlineCents
	}
	if product == nil {
		return booking.TotalPriceCents
	}
	amount, err := catalogDomain.AmountDueOnline(product.Deposit, booking.TotalPriceCents)
	if err != nil {
		return booking.TotalPriceCents
	}
	return amount
}

// loadBooking reads and decrypts a booking.
func (s *BookingService) loadBooking(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	bookingEncx, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, errs.ErrRepositoryNotFound
		}
		return nil, fmt.Errorf("get booking: %w", err)
	}
	booking, err := domain.DecryptBookingEncx(ctx, s.crypto, bookingEncx)
	if err != nil {
		return nil, fmt.Errorf("decrypt booking: %w", err)
	}
	return booking, nil
}

// bookingProduct reads the booking's product; nil when it no longer exists
// (the cancellation then refunds by the default rule).
func (s *BookingService) bookingProduct(ctx context.Context, booking *domain.Booking) (*catalogDomain.ProductRes, error) {
	if s.productService == nil {
		return nil, nil
	}
	product, err := s.productService.GetProductByID(ctx, booking.ProductID.String())
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get product of booking: %w", err)
	}
	return product, nil
}

// releaseAvailability marks the booking's availability as available again
// (best effort).
func (s *BookingService) releaseAvailability(ctx context.Context, booking *domain.Booking) {
	availabilityEncx, err := s.availabilityRepo.GetByID(ctx, booking.AvailabilityID)
	if err != nil {
		return
	}
	availability, err := domain.DecryptAvailabilityEncx(ctx, s.crypto, availabilityEncx)
	if err != nil {
		return
	}
	availability.MarkAsAvailable()
	if updatedEncx, err := domain.ProcessAvailabilityEncx(ctx, s.crypto, availability); err == nil {
		_ = s.availabilityRepo.Update(ctx, updatedEncx)
	}
}

// sendCancellationNotice emails the client and the partner that the booking
// was cancelled, with the refund (best effort).
func (s *BookingService) sendCancellationNotice(ctx context.Context, booking *domain.Booking, product *catalogDomain.ProductRes) {
	if s.notificationService == nil {
		return
	}
	productName := ""
	if product != nil {
		productName = product.Name
	}
	data := s.buildNotificationData(booking, productName)
	data.AmountPaidOnlineCents = amountPaidOnline(booking, product)
	if booking.RefundedAmountCents != nil {
		data.RefundedAmountCents = *booking.RefundedAmountCents
	}
	if err := s.notificationService.SendBookingCancellation(ctx, data); err != nil {
		slog.WarnContext(ctx, "failed to send booking cancellation notification",
			"booking_id", booking.ID,
			"err", err,
		)
	}
}
