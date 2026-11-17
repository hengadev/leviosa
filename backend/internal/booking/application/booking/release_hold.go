package booking

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

// errPaymentUnderWay refuses releasing a hold whose PaymentIntent Stripe
// would not cancel: the payment is processing or went through, and its
// webhook settles the booking.
var errPaymentUnderWay = errors.New("a payment is under way on this booking")

// ReleaseHoldByToken ends the payment hold of a booking the client walked
// away from (ADR-0018): a declined card followed by "change slot", a closed
// page, a failed redirect payment. Without it the slot stayed taken until the
// hold elapsed. The PaymentIntent is canceled before the booking is expired,
// so a payment can never reach a released booking; its canceled webhook is
// only logged (HandlePaymentWebhook).
func (s *BookingService) ReleaseHoldByToken(ctx context.Context, token string) error {
	if len(s.tokenSecret) == 0 {
		return fmt.Errorf("booking token feature is not configured")
	}

	bookingID, err := domain.VerifyBookingToken(token, s.tokenSecret)
	if err != nil {
		if domain.IsBookingTokenError(err) {
			return errs.NewUnauthorizedErr(err.Error())
		}
		return fmt.Errorf("verify booking token: %w", err)
	}

	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return errs.NewUnauthorizedErr("booking not found")
		}
		return fmt.Errorf("get booking to release its hold: %w", err)
	}
	if booking.Status != domain.BookingStatusAwaitingPayment {
		// Confirmed, expired or cancelled already: nothing is held.
		return nil
	}

	if booking.PaymentIntentID != nil && *booking.PaymentIntentID != "" {
		if err := s.paymentService.CancelPaymentIntent(ctx, *booking.PaymentIntentID); err != nil {
			slog.WarnContext(ctx, "payment hold kept: its payment intent could not be canceled",
				"booking_id", booking.ID, "error", err)
			return errs.NewConflictErr(errPaymentUnderWay)
		}
	}

	released, err := s.bookingRepo.ReleaseHold(ctx, booking.ID, s.now())
	if err != nil {
		return fmt.Errorf("release payment hold: %w", err)
	}
	if released {
		slog.InfoContext(ctx, "payment hold released by the client", "booking_id", booking.ID)
	}
	return nil
}
