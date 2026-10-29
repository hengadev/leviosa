package booking

import (
	"context"
	"net/http"
	"testing"
	"time"

	bookingService "github.com/Leviosa-care/leviosa/backend/internal/booking/application/booking"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	notificationDomain "github.com/Leviosa-care/leviosa/backend/internal/notification/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Payment hold expiry and late payments (ADR-0018,
// docs/issues/015-payment-hold-expiry.md). Stripe is faked at the payment
// port (MockPaymentService).
//
// make test-func TEST_NAME=TestPaymentHoldExpiry TEST_PATH=test/integration/booking/booking/hold_expiry_test.go

func mockPayments(t *testing.T) *MockPaymentService {
	t.Helper()
	mock, ok := paymentService.(*MockPaymentService)
	require.True(t, ok, "the suite fakes Stripe with MockPaymentService")
	return mock
}

// elapseHold moves a booking's payment hold into the past, as if the client
// had left the payment step 15 minutes ago.
func elapseHold(t *testing.T, ctx context.Context, bookingID uuid.UUID) {
	t.Helper()
	_, err := testPool.Exec(ctx, `UPDATE booking.bookings SET hold_expires_at = NOW() - INTERVAL '1 minute' WHERE id = $1`, bookingID)
	require.NoError(t, err)
}

func storedBooking(t *testing.T, ctx context.Context, bookingID uuid.UUID) *domain.Booking {
	t.Helper()
	bookingEncx, err := bookingRepo.GetByID(ctx, bookingID)
	require.NoError(t, err)
	booking, err := domain.DecryptBookingEncx(ctx, crypto, bookingEncx)
	require.NoError(t, err)
	return booking
}

func holdGuestBooking(t *testing.T, ctx context.Context, f slotFixture, start time.Time, email string, locale domain.BookingLocale) *domain.Booking {
	t.Helper()
	booking, err := service.CreateBooking(ctx, f.availabilityID, nil, f.productID, start, "Claire", "Dupont", email, "0611112222", locale)
	require.NoError(t, err)
	require.Equal(t, domain.BookingStatusAwaitingPayment, booking.Status)
	require.NotNil(t, booking.PaymentIntentID)
	return booking
}

func slotsOfDay(t *testing.T, client *http.Client, f slotFixture) []time.Time {
	t.Helper()
	day := f.day.Format(time.DateOnly)
	status, body := getFreeSlots(t, client, f.partnerID, f.productID, day, day)
	require.Equal(t, http.StatusOK, status)
	return slotStarts(body)
}

func confirmations(s *SpyNotificationService) []ports.BookingNotificationData {
	return s.BookingConfirmations
}

func refundNotices(s *SpyNotificationService) []ports.BookingNotificationData {
	return s.LatePaymentRefunds
}

func paymentFailures(s *SpyNotificationService) []ports.BookingNotificationData {
	return s.PaymentFailed
}

func cancellations(s *SpyNotificationService) []ports.BookingNotificationData {
	return s.BookingCancellations
}

func TestPaymentHoldExpiry(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}
	payments := mockPayments(t)
	job := bookingService.NewHoldExpiryJob(service, 0)

	t.Run("should expire an unpaid hold and offer its slot again", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)

		held := holdGuestBooking(t, ctx, f, f.at(10, 15), "claire@example.com", domain.BookingLocaleFR)
		running := holdGuestBooking(t, ctx, f, f.at(11, 30), "other@example.com", domain.BookingLocaleFR)
		elapseHold(t, ctx, held.ID)

		job.TickOnce(ctx)

		assert.Equal(t, domain.BookingStatusExpired, storedBooking(t, ctx, held.ID).Status)
		assert.Equal(t, 1, payments.CancelsOf(*held.PaymentIntentID), "the expired hold's PaymentIntent is canceled")
		assert.Equal(t, ports.PaymentIntentStatusCanceled, payments.IntentStatus(*held.PaymentIntentID))

		assert.Equal(t, domain.BookingStatusAwaitingPayment, storedBooking(t, ctx, running.ID).Status, "a running hold is left alone")
		assert.Equal(t, 0, payments.CancelsOf(*running.PaymentIntentID))

		assertSameInstants(t, []time.Time{f.at(9, 0), f.at(10, 15)}, slotsOfDay(t, client, f))
	})

	t.Run("should change nothing more when the job runs twice", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)

		held := holdGuestBooking(t, ctx, f, f.at(10, 15), "claire@example.com", domain.BookingLocaleFR)
		elapseHold(t, ctx, held.ID)

		expired, err := service.ExpireElapsedHolds(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, expired)
		first := storedBooking(t, ctx, held.ID)

		expired, err = service.ExpireElapsedHolds(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, expired, "the second run expires nothing")

		second := storedBooking(t, ctx, held.ID)
		assert.Equal(t, domain.BookingStatusExpired, second.Status)
		assert.True(t, first.UpdatedAt.Equal(second.UpdatedAt), "the booking is not written again")
		assert.Equal(t, 1, payments.CancelsOf(*held.PaymentIntentID), "the PaymentIntent is canceled once")
	})

	t.Run("should leave a booking confirmed before the job runs", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)

		held := holdGuestBooking(t, ctx, f, f.at(10, 15), "claire@example.com", domain.BookingLocaleFR)
		confirmBookingPayment(t, ctx, held.ID)
		elapseHold(t, ctx, held.ID)

		job.TickOnce(ctx)

		assert.Equal(t, domain.BookingStatusConfirmed, storedBooking(t, ctx, held.ID).Status)
		assert.Equal(t, 0, payments.CancelsOf(*held.PaymentIntentID))
	})

	t.Run("should send no email and succeed when the canceled webhook arrives for an expired booking", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)

		held := holdGuestBooking(t, ctx, f, f.at(10, 15), "claire@example.com", domain.BookingLocaleFR)
		elapseHold(t, ctx, held.ID)
		job.TickOnce(ctx)

		event := paymentSucceededEvent(t, ctx, held.ID)
		event.Type = ports.WebhookEventPaymentIntentCanceled
		event.Status = ports.PaymentIntentStatusCanceled
		require.NoError(t, service.HandlePaymentWebhook(ctx, event))
		require.NoError(t, service.HandlePaymentWebhook(ctx, event))

		stored := storedBooking(t, ctx, held.ID)
		assert.Equal(t, domain.BookingStatusExpired, stored.Status)
		assert.Equal(t, 0, notificationSpy.CountFor(paymentFailures, held.ID), "no payment-failed email")
		assert.Equal(t, 0, notificationSpy.CountFor(cancellations, held.ID), "no cancellation email")
	})

	t.Run("should send no email when a payment attempt fails during the hold", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)

		held := holdGuestBooking(t, ctx, f, f.at(10, 15), "claire@example.com", domain.BookingLocaleFR)

		event := paymentSucceededEvent(t, ctx, held.ID)
		event.Type = ports.WebhookEventPaymentIntentPaymentFailed
		event.Status = ports.PaymentIntentStatusRequiresPaymentMethod
		event.FailureCode = "card_declined"
		require.NoError(t, service.HandlePaymentWebhook(ctx, event))

		assert.Equal(t, domain.BookingStatusAwaitingPayment, storedBooking(t, ctx, held.ID).Status)
		assert.Equal(t, 0, notificationSpy.CountFor(paymentFailures, held.ID), "the client retries on the page")

		// The client retries and pays: the booking is confirmed.
		confirmBookingPayment(t, ctx, held.ID)
	})
}

func TestLatePayment(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}
	payments := mockPayments(t)
	job := bookingService.NewHoldExpiryJob(service, 0)

	// Stripe reports the payment of an intent the job tried to cancel: the
	// client paid just before the cancellation.
	paidLate := func(t *testing.T, booking *domain.Booking, amount int) *ports.WebhookEvent {
		t.Helper()
		require.NoError(t, payments.ConfirmPaymentIntent(ctx, *booking.PaymentIntentID))
		event := paymentSucceededEvent(t, ctx, booking.ID)
		event.Amount = amount
		return event
	}

	t.Run("should confirm an expired booking whose slot is still free", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)

		held := holdGuestBooking(t, ctx, f, f.at(10, 15), "claire@example.com", domain.BookingLocaleFR)
		elapseHold(t, ctx, held.ID)
		job.TickOnce(ctx)
		require.Equal(t, domain.BookingStatusExpired, storedBooking(t, ctx, held.ID).Status)

		event := paidLate(t, held, 1500)
		require.NoError(t, service.HandlePaymentWebhook(ctx, event))
		require.NoError(t, service.HandlePaymentWebhook(ctx, event), "a duplicate delivery succeeds")

		stored := storedBooking(t, ctx, held.ID)
		assert.Equal(t, domain.BookingStatusConfirmed, stored.Status)
		assert.Equal(t, domain.PaymentStatusPaid, stored.PaymentStatus)
		assert.Equal(t, 1, notificationSpy.CountFor(confirmations, held.ID), "one booking-confirmed event")
		assert.Empty(t, payments.RefundsFor(*held.PaymentIntentID))
		assert.Equal(t, 0, notificationSpy.CountFor(refundNotices, held.ID))

		// The slot is taken again.
		assertSameInstants(t, []time.Time{f.at(9, 0), f.at(11, 30)}, slotsOfDay(t, client, f))
		assert.Equal(t, http.StatusConflict, postBooking(t, client, f.availabilityID, f.productID, f.at(10, 15)))
	})

	t.Run("should confirm a booking whose hold elapsed before the job ran, its slot still free", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)

		held := holdGuestBooking(t, ctx, f, f.at(10, 15), "claire@example.com", domain.BookingLocaleFR)
		elapseHold(t, ctx, held.ID)

		require.NoError(t, service.HandlePaymentWebhook(ctx, paidLate(t, held, 1500)))

		assert.Equal(t, domain.BookingStatusConfirmed, storedBooking(t, ctx, held.ID).Status)
		assert.Equal(t, 1, notificationSpy.CountFor(confirmations, held.ID))
	})

	t.Run("should refund in full a late payment whose slot was taken", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)

		notificationEmailSpy.Reset()
		stop := notificationSpy.ForwardConfirmationsTo(realNotificationAdapter)
		defer stop()

		late := holdGuestBooking(t, ctx, f, f.at(10, 15), "late@example.com", domain.BookingLocaleEN)
		elapseHold(t, ctx, late.ID)
		// Another client books the freed slot, and pays.
		other := holdGuestBooking(t, ctx, f, f.at(10, 15), "other@example.com", domain.BookingLocaleFR)
		confirmBookingPayment(t, ctx, other.ID)
		job.TickOnce(ctx)
		require.Equal(t, domain.BookingStatusExpired, storedBooking(t, ctx, late.ID).Status)

		const paidCents = 1500
		event := paidLate(t, late, paidCents)
		require.NoError(t, service.HandlePaymentWebhook(ctx, event))
		require.NoError(t, service.HandlePaymentWebhook(ctx, event), "a duplicate delivery succeeds")

		refunds := payments.RefundsFor(*late.PaymentIntentID)
		require.Len(t, refunds, 1, "refunded once")
		assert.Equal(t, paidCents, refunds[0].Amount, "refunded in full: the amount paid")

		stored := storedBooking(t, ctx, late.ID)
		assert.Equal(t, domain.BookingStatusExpired, stored.Status, "the booking stays expired")
		assert.Equal(t, domain.PaymentStatusRefunded, stored.PaymentStatus, "the refund is recorded")
		assert.Contains(t, stored.PartnerNotes, "Late payment refunded")
		assert.Equal(t, domain.BookingStatusConfirmed, storedBooking(t, ctx, other.ID).Status, "the other client keeps the slot")

		assert.Equal(t, 0, notificationSpy.CountFor(confirmations, late.ID), "no confirmation for the late booking")
		assert.Equal(t, 1, notificationSpy.CountFor(refundNotices, late.ID))

		require.Eventually(t, func() bool {
			return len(notificationEmailSpy.LatePaymentRefundSnapshot()) >= 1
		}, 5*time.Second, 20*time.Millisecond, "expected the refund email")
		time.Sleep(200 * time.Millisecond) // give a duplicate a chance to show up
		emails := notificationEmailSpy.LatePaymentRefundSnapshot()
		require.Len(t, emails, 1, "exactly one refund email")
		assert.Equal(t, "late@example.com", emails[0].ToEmail)
		assert.Equal(t, notificationDomain.LocaleEN, emails[0].Locale, "in the booking's locale")
		assert.Equal(t, paidCents, emails[0].RefundedAmountCents)
		assert.True(t, emails[0].StartTime.Equal(late.SlotStartTime))
		for _, confirmation := range func() []notificationDomain.BookingConfirmationRequest {
			clients, _ := notificationEmailSpy.Snapshot()
			return clients
		}() {
			assert.NotEqual(t, late.ID.String(), confirmation.BookingID, "no confirmation email for the late booking")
		}
	})

	t.Run("should expire and refund a booking whose hold elapsed before the job ran, its slot taken", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)

		late := holdGuestBooking(t, ctx, f, f.at(10, 15), "late@example.com", domain.BookingLocaleFR)
		elapseHold(t, ctx, late.ID)
		other := holdGuestBooking(t, ctx, f, f.at(10, 15), "other@example.com", domain.BookingLocaleFR)

		require.NoError(t, service.HandlePaymentWebhook(ctx, paidLate(t, late, 1500)))

		stored := storedBooking(t, ctx, late.ID)
		assert.Equal(t, domain.BookingStatusExpired, stored.Status)
		assert.Equal(t, domain.PaymentStatusRefunded, stored.PaymentStatus)
		require.Len(t, payments.RefundsFor(*late.PaymentIntentID), 1)
		assert.Equal(t, 0, notificationSpy.CountFor(confirmations, late.ID))
		assert.Equal(t, 1, notificationSpy.CountFor(refundNotices, late.ID))
		assert.Equal(t, domain.BookingStatusAwaitingPayment, storedBooking(t, ctx, other.ID).Status, "the other client's hold is untouched")
	})
}
