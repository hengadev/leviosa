package bookingHelpers

import (
	"context"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// PaymentSucceededEvent is the payment_intent.succeeded event Stripe sends
// for a booking's PaymentIntent (Stripe is faked at the port).
func PaymentSucceededEvent(t *testing.T, ctx context.Context, repo ports.BookingRepository, bookingID uuid.UUID) *ports.WebhookEvent {
	t.Helper()
	bookingEncx, err := repo.GetByID(ctx, bookingID)
	require.NoError(t, err)
	require.NotNil(t, bookingEncx.PaymentIntentID, "the booking has a PaymentIntent")
	return &ports.WebhookEvent{
		ID:              "evt_test_" + uuid.NewString(),
		Type:            ports.WebhookEventPaymentIntentSucceeded,
		PaymentIntentID: *bookingEncx.PaymentIntentID,
		Status:          "succeeded",
		Amount:          bookingEncx.TotalPriceCents,
		Currency:        bookingEncx.Currency,
		Metadata:        map[string]string{"booking_id": bookingID.String()},
	}
}

// ConfirmPayment delivers the payment-succeeded webhook for a booking
// awaiting payment, as Stripe does once the client paid: the booking becomes
// confirmed and paid (ADR-0018).
func ConfirmPayment(t *testing.T, ctx context.Context, svc ports.BookingService, repo ports.BookingRepository, bookingID uuid.UUID) {
	t.Helper()
	require.NoError(t, svc.HandlePaymentWebhook(ctx, PaymentSucceededEvent(t, ctx, repo, bookingID)))

	bookingEncx, err := repo.GetByID(ctx, bookingID)
	require.NoError(t, err)
	require.Equal(t, domain.BookingStatusConfirmed, bookingEncx.Status)
	require.Equal(t, domain.PaymentStatusPaid, bookingEncx.PaymentStatus)
}
