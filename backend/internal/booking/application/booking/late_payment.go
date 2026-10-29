package booking

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/Leviosa-care/leviosa/backend/internal/common/audit"

	"github.com/google/uuid"
)

// Payments on bookings that are not appointments yet (ADR-0018,
// docs/issues/015-payment-hold-expiry.md). A payment that succeeds while the
// hold runs confirms the booking: the hold kept the slot. One that succeeds
// after the hold elapsed (the booking still awaiting payment, or already
// expired by the job) confirms it only if the slot is still free, checked
// under the partner's agenda lock so no booking can take it meanwhile;
// otherwise it is refunded in full and the client is told why.

// errLatePaymentSlotTaken refuses confirming a late payment whose slot was
// booked by someone else after the hold elapsed.
var errLatePaymentSlotTaken = errors.New("the slot of the late payment is no longer free")

// settleHeldPayment handles payment_intent.succeeded on a booking awaiting
// payment or expired. Duplicate deliveries confirm or refund once.
func (s *BookingService) settleHeldPayment(ctx context.Context, booking *domain.Booking, event *ports.WebhookEvent) error {
	buffers := map[uuid.UUID]time.Duration{}
	buffer := s.bufferOf(ctx, booking.ProductID, buffers, 0)
	wasExpired := booking.Status == domain.BookingStatusExpired

	confirmed, err := s.bookingRepo.ConfirmGuarded(ctx, booking.ID, booking.PartnerID, func(current *domain.BookingEncx, others []*domain.BookingEncx) error {
		// Read the clock under the lock: a booking created before it saw
		// this hold either running (and was refused) or elapsed (and is
		// among others).
		now := time.Now()
		if current.Status == domain.BookingStatusAwaitingPayment &&
			(current.HoldExpiresAt == nil || now.Before(*current.HoldExpiresAt)) {
			return nil // the hold still keeps the slot
		}
		occupancies, err := s.occupancies(ctx, others, buffers, buffer, now)
		if err != nil {
			return err
		}
		if booking.OverlapsAny(buffer, occupancies) {
			return errLatePaymentSlotTaken
		}
		return nil
	})
	switch {
	case errors.Is(err, errLatePaymentSlotTaken):
		return s.refundLatePayment(ctx, booking, event)
	case err != nil:
		return fmt.Errorf("confirm booking after payment: %w", err)
	case !confirmed:
		// Another delivery of the event settled it already.
		return nil
	}

	_ = booking.ConfirmLatePayment()
	_ = booking.MarkPaymentPaid()
	// What the client paid online, for the confirmation email (the amount
	// due online is not persisted).
	booking.AmountDueOnlineCents = event.Amount
	if wasExpired {
		slog.InfoContext(ctx, "late payment confirmed an expired booking, its slot still free",
			"booking_id", booking.ID, "stripe_event_id", event.ID)
	}

	productName := ""
	if s.productService != nil {
		productName = s.resolveProductName(ctx, booking.ProductID)
	}
	s.sendBookingConfirmed(ctx, booking, productName)
	return nil
}

// refundLatePayment refunds in full a payment that arrived after the hold,
// for a slot someone else booked meanwhile. The booking stays (or becomes)
// expired with its payment refunded, and the client is emailed. The refund
// is claimed first, so a duplicate delivery refunds nothing; a failed
// refund releases the claim and fails the webhook, so Stripe retries.
func (s *BookingService) refundLatePayment(ctx context.Context, booking *domain.Booking, event *ports.WebhookEvent) error {
	if booking.Status == domain.BookingStatusAwaitingPayment {
		if _, err := s.bookingRepo.ExpireHold(ctx, booking.ID, time.Now()); err != nil {
			return fmt.Errorf("expire booking before refunding its late payment: %w", err)
		}
	}

	claimed, err := s.bookingRepo.ClaimLatePaymentRefund(ctx, booking.ID)
	if err != nil {
		return fmt.Errorf("claim late payment refund: %w", err)
	}
	if !claimed {
		slog.InfoContext(ctx, "late payment already refunded, or its booking is not expired",
			"booking_id", booking.ID, "stripe_event_id", event.ID)
		return nil
	}

	slog.WarnContext(ctx, "late payment on a slot taken meanwhile: refunding in full",
		"booking_id", booking.ID, "stripe_event_id", event.ID, "amount_cents", event.Amount)

	refundID, err := s.paymentService.RefundPayment(ctx, *booking.PaymentIntentID, event.Amount, "")
	if err != nil {
		if releaseErr := s.bookingRepo.ReleaseLatePaymentRefund(ctx, booking.ID); releaseErr != nil {
			slog.ErrorContext(ctx, "failed to release the late payment refund claim",
				"booking_id", booking.ID, "error", releaseErr)
		}
		return fmt.Errorf("refund late payment: %w", err)
	}

	booking.Status = domain.BookingStatusExpired
	booking.PaymentStatus = domain.PaymentStatusRefunded
	s.recordLatePaymentRefund(ctx, booking.ID, refundID, event.Amount)

	audit.WriteOrLog(ctx, s.audit, audit.Entry{
		ActorRole:  "system",
		Action:     "booking.late_payment_refunded",
		TargetType: "booking",
		TargetID:   booking.ID.String(),
		Payload: map[string]any{
			"refund_id":         refundID,
			"payment_intent_id": *booking.PaymentIntentID,
			"amount_cents":      event.Amount,
			"stripe_event_id":   event.ID,
		},
	})

	if s.latePaymentNotifier != nil {
		data := s.buildNotificationData(booking, "")
		if s.productService != nil {
			data.ProductName = s.resolveProductName(ctx, booking.ProductID)
		}
		data.RefundedAmountCents = event.Amount
		if err := s.latePaymentNotifier.SendLatePaymentRefund(ctx, data); err != nil {
			slog.ErrorContext(ctx, "failed to send late payment refund notification",
				"booking_id", booking.ID, "error", err)
		}
	}
	return nil
}

// recordLatePaymentRefund keeps the refund on the booking, in the internal
// bookkeeping notes as RefundBooking does. Best effort: the money is already
// back and the payment status says so.
func (s *BookingService) recordLatePaymentRefund(ctx context.Context, bookingID uuid.UUID, refundID string, amountCents int) {
	bookingEncx, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err == nil {
		var booking *domain.Booking
		booking, err = domain.DecryptBookingEncx(ctx, s.crypto, bookingEncx)
		if err == nil {
			booking.SetPartnerNotes(fmt.Sprintf("Late payment refunded in full (slot no longer available): %s, %d cents", refundID, amountCents))
			bookingEncx, err = domain.ProcessBookingEncx(ctx, s.crypto, booking)
			if err == nil {
				err = s.bookingRepo.Update(ctx, bookingEncx)
			}
		}
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to record the late payment refund on the booking",
			"booking_id", bookingID, "refund_id", refundID, "error", err)
	}
}
