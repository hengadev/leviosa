package booking

import (
	"context"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	tbooking "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/booking"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// paymentSucceededEvent is the payment_intent.succeeded event Stripe sends
// for a booking's PaymentIntent.
func paymentSucceededEvent(t *testing.T, ctx context.Context, bookingID uuid.UUID) *ports.WebhookEvent {
	t.Helper()
	return tbooking.PaymentSucceededEvent(t, ctx, bookingRepo, bookingID)
}

// confirmBookingPayment has Stripe confirm a booking awaiting payment
// (ADR-0018). Tests about confirmed bookings call it right after creating one.
// The faked PaymentIntent succeeds too, as Stripe's does, so the booking's
// payment can be refunded.
func confirmBookingPayment(t *testing.T, ctx context.Context, bookingID uuid.UUID) {
	t.Helper()
	event := paymentSucceededEvent(t, ctx, bookingID)
	require.NoError(t, paymentService.ConfirmPaymentIntent(ctx, event.PaymentIntentID))
	tbooking.ConfirmPayment(t, ctx, service, bookingRepo, bookingID)
}

// practiceToday is today's civil date at the practice (Europe/Paris), as a
// UTC midnight: the form the financial summary takes its bounds in. Taking
// the UTC date instead puts a payment made between midnight in Paris and
// midnight UTC on the wrong day.
func practiceToday() time.Time {
	now := time.Now().In(domain.PracticeLocation)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

// monthSoFar is the financial summary's window for this month at the
// practice, up to and including today.
func monthSoFar() (from, to time.Time) {
	today := practiceToday()
	return time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC), today.AddDate(0, 0, 1)
}
