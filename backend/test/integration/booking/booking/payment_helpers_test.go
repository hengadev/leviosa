package booking

import (
	"context"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	tbooking "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/booking"

	"github.com/google/uuid"
)

// paymentSucceededEvent is the payment_intent.succeeded event Stripe sends
// for a booking's PaymentIntent.
func paymentSucceededEvent(t *testing.T, ctx context.Context, bookingID uuid.UUID) *ports.WebhookEvent {
	t.Helper()
	return tbooking.PaymentSucceededEvent(t, ctx, bookingRepo, bookingID)
}

// confirmBookingPayment has Stripe confirm a booking awaiting payment
// (ADR-0018). Tests about confirmed bookings call it right after creating one.
func confirmBookingPayment(t *testing.T, ctx context.Context, bookingID uuid.UUID) {
	t.Helper()
	tbooking.ConfirmPayment(t, ctx, service, bookingRepo, bookingID)
}
