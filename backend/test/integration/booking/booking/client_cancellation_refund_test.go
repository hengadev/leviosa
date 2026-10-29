package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	notificationDomain "github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
	tsetup "github.com/Leviosa-care/leviosa/backend/test/helpers/booking"
	tavail "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/availability"
	tbooking "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/booking"
	tb "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/building"
	tr "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/room"
	tcatalog "github.com/Leviosa-care/leviosa/backend/test/helpers/catalog"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// make test-func TEST_NAME=TestClientCancellationRefund TEST_PATH=test/integration/booking/booking/client_cancellation_refund_test.go

// The product of these tests: a fixed 30 € deposit on an 80 € price, and the
// ADR-0014 example tiers.
const (
	refundTestDepositCents = 3000
	refundTestTiers        = `[{"minHoursBefore":48,"refundPercentage":100},{"minHoursBefore":24,"refundPercentage":50},{"minHoursBefore":0,"refundPercentage":0}]`
)

// setupTieredBooking books, as an English-speaking guest, the product with a
// deposit and tiers, slotIn from now. When paid, Stripe confirms the deposit
// payment. It returns the booking (with its token).
func setupTieredBooking(t *testing.T, ctx context.Context, slotIn time.Duration, paid bool) *domain.Booking {
	t.Helper()

	tbooking.ClearBookingsTable(t, ctx, testPool)
	tavail.ClearAvailabilityTable(t, ctx, testPool)
	tr.ClearRoomsTable(t, ctx, testPool)
	tb.ClearBuildingsTable(t, ctx, testPool)
	tcatalog.ClearProductsTable(t, ctx, testPool)

	categoryID := tcatalog.CreateTestCategory(t, ctx, testPool)
	products := tcatalog.CreateDefaultTestProducts(t, ctx, testPool, categoryID)
	productID := products[0].ID
	_, err := testPool.Exec(ctx, `UPDATE catalog.products SET deposit = $2, cancellation_tiers = $3 WHERE id = $1`,
		productID, fmt.Sprintf(`{"type":"fixed","value":%d}`, refundTestDepositCents), refundTestTiers)
	require.NoError(t, err)

	building := tb.NewTestBuilding(t)
	buildingEncx, err := domain.ProcessBuildingEncx(ctx, crypto, building)
	require.NoError(t, err)
	require.NoError(t, tb.InsertBuildingEncx(t, ctx, testPool, buildingEncx))
	room := tr.NewTestRoomWithBuilding(t, building.ID)
	roomEncx, err := domain.ProcessRoomEncx(ctx, crypto, room)
	require.NoError(t, err)
	require.NoError(t, tr.InsertRoomEncx(t, ctx, testPool, roomEncx))

	_, partnerID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "partner-refund@example.com", room.ID, testPool, authCtx.Redis, crypto)

	start := time.Now().Add(slotIn).Truncate(10 * time.Minute).Add(10 * time.Minute)
	availability := tavail.NewTestAvailability(t, partnerID, room.ID, start, start.Add(2*time.Hour), 0)
	availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
	require.NoError(t, err)
	tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)

	booking, err := service.CreateBooking(ctx, availability.ID, nil, productID, start,
		"Jane", "Doe", "jane.refund@example.com", "+33612345678", domain.BookingLocaleEN)
	require.NoError(t, err)
	require.Equal(t, refundTestDepositCents, booking.AmountDueOnlineCents)

	if paid {
		event := paymentSucceededEvent(t, ctx, booking.ID)
		event.Amount = refundTestDepositCents
		require.NoError(t, paymentService.ConfirmPaymentIntent(ctx, event.PaymentIntentID))
		require.NoError(t, service.HandlePaymentWebhook(ctx, event))
	}
	return booking
}

func lookupByToken(t *testing.T, token string) domain.PublicBookingLookupResponse {
	t.Helper()
	resp, err := http.Get(testServerURL + "/bookings/lookup?token=" + token)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var lookup domain.PublicBookingLookupResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&lookup))
	return lookup
}

// cancelByToken posts the client cancellation from the booking link, with no
// body (no reason is collected, ADR-0019).
func cancelByToken(t *testing.T, bookingID uuid.UUID, token string) (int, domain.PublicBookingLookupResponse) {
	t.Helper()
	url := fmt.Sprintf("%s/bookings/%s/cancel-public?token=%s", testServerURL, bookingID, token)
	resp, err := http.Post(url, "", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	var result domain.PublicBookingLookupResponse
	if resp.StatusCode == http.StatusOK {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	}
	return resp.StatusCode, result
}

func TestClientCancellationRefund(t *testing.T) {
	ctx := context.Background()
	payments := mockPayments(t)

	t.Run("refunds the tier amount, as previewed, and emails client and partner once", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		// 30 hours ahead: the 24 h tier, 50 % of the deposit.
		booking := setupTieredBooking(t, ctx, 30*time.Hour, true)

		stored, err := bookingRepo.GetByID(ctx, booking.ID)
		require.NoError(t, err)
		require.NotNil(t, stored.AmountPaidOnlineCents, "the payment webhook records the amount paid online")
		assert.Equal(t, refundTestDepositCents, *stored.AmountPaidOnlineCents)

		lookup := lookupByToken(t, booking.Token)
		assert.Equal(t, domain.BookingLocaleEN, lookup.Locale)
		assert.Equal(t, refundTestDepositCents, lookup.AmountPaidOnlineCents)
		require.NotNil(t, lookup.RefundPreview, "a confirmed booking shows the refund it would get now")
		assert.Equal(t, refundTestDepositCents/2, lookup.RefundPreview.AmountCents)
		assert.Equal(t, 50, lookup.RefundPreview.Percentage)
		require.NotNil(t, lookup.RefundPreview.MinHoursBefore)
		assert.Equal(t, 24, *lookup.RefundPreview.MinHoursBefore)

		notificationEmailSpy.Reset()
		smsBefore := notificationSMSSpy.Count()
		stop := notificationSpy.ForwardConfirmationsTo(realNotificationAdapter)
		defer stop()

		status, result := cancelByToken(t, booking.ID, booking.Token)
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, domain.BookingStatusCancelled, result.Status)
		assert.Equal(t, domain.CancelledByClient, result.CancelledBy)
		require.NotNil(t, result.RefundedAmountCents)
		assert.Equal(t, lookup.RefundPreview.AmountCents, *result.RefundedAmountCents, "the refund made is the one previewed")
		assert.Equal(t, domain.PaymentStatusPartiallyRefunded, result.PaymentStatus)
		assert.Nil(t, result.RefundPreview, "a cancelled booking has nothing left to preview")

		// The payment port refunded the tier amount, once.
		refunds := payments.RefundsFor(*booking.PaymentIntentID)
		require.Len(t, refunds, 1)
		assert.Equal(t, lookup.RefundPreview.AmountCents, refunds[0].Amount)

		// The booking records who cancelled and what was refunded.
		stored, err = bookingRepo.GetByID(ctx, booking.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.BookingStatusCancelled, stored.Status)
		assert.Equal(t, domain.PaymentStatusPartiallyRefunded, stored.PaymentStatus)
		require.NotNil(t, stored.CancelledBy)
		assert.Equal(t, "client", *stored.CancelledBy)
		require.NotNil(t, stored.RefundedAmountCents)
		assert.Equal(t, refundTestDepositCents/2, *stored.RefundedAmountCents)
		assert.NotNil(t, stored.CancelledAt)

		// One cancellation email to the client, in the booking's locale, and
		// one to the partner; no SMS.
		require.Eventually(t, func() bool {
			clients, partners := notificationEmailSpy.CancellationSnapshot()
			return len(clients) >= 1 && len(partners) >= 1
		}, 5*time.Second, 20*time.Millisecond, "expected the client and partner cancellation emails")

		// A second cancel request is refused and refunds nothing more.
		status, _ = cancelByToken(t, booking.ID, booking.Token)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Len(t, payments.RefundsFor(*booking.PaymentIntentID), 1, "no second refund")

		time.Sleep(200 * time.Millisecond) // give a duplicate email a chance to show up
		clients, partners := notificationEmailSpy.CancellationSnapshot()
		require.Len(t, clients, 1, "exactly one client cancellation email")
		require.Len(t, partners, 1, "exactly one partner cancellation email")

		client := clients[0]
		assert.Equal(t, "jane.refund@example.com", client.ToEmail)
		assert.Equal(t, notificationDomain.LocaleEN, client.Locale)
		assert.Equal(t, notificationDomain.CancelledByClient, client.CancelledBy)
		assert.Equal(t, refundTestDepositCents, client.AmountPaidOnlineCents)
		assert.Equal(t, refundTestDepositCents/2, client.RefundedAmountCents)
		assert.True(t, client.StartTime.Equal(booking.SlotStartTime))

		partner := partners[0]
		assert.Equal(t, "partner-refund@example.com", partner.ToEmail)
		assert.Equal(t, "Jane Doe", partner.ClientName)
		assert.Equal(t, refundTestDepositCents/2, partner.RefundedAmountCents)

		assert.Equal(t, smsBefore, notificationSMSSpy.Count(), "booking notifications send no SMS")
	})

	t.Run("a tier refunding nothing makes no refund call", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		// 10 hours ahead: the 0 h tier, nothing refunded.
		booking := setupTieredBooking(t, ctx, 10*time.Hour, true)

		lookup := lookupByToken(t, booking.Token)
		require.NotNil(t, lookup.RefundPreview)
		assert.Equal(t, 0, lookup.RefundPreview.AmountCents)
		assert.Equal(t, 0, lookup.RefundPreview.Percentage)

		status, result := cancelByToken(t, booking.ID, booking.Token)
		require.Equal(t, http.StatusOK, status)
		require.NotNil(t, result.RefundedAmountCents)
		assert.Equal(t, 0, *result.RefundedAmountCents)
		assert.Equal(t, domain.PaymentStatusPaid, result.PaymentStatus, "nothing refunded: the payment stays paid")
		assert.Empty(t, payments.RefundsFor(*booking.PaymentIntentID), "a zero refund calls no refund")
	})

	t.Run("well ahead the whole deposit is refunded", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		booking := setupTieredBooking(t, ctx, 72*time.Hour, true)

		status, result := cancelByToken(t, booking.ID, booking.Token)
		require.Equal(t, http.StatusOK, status)
		require.NotNil(t, result.RefundedAmountCents)
		assert.Equal(t, refundTestDepositCents, *result.RefundedAmountCents)
		assert.Equal(t, domain.PaymentStatusRefunded, result.PaymentStatus)
		refunds := payments.RefundsFor(*booking.PaymentIntentID)
		require.Len(t, refunds, 1)
		assert.Equal(t, refundTestDepositCents, refunds[0].Amount)
	})

	t.Run("a booking not confirmed cannot be cancelled", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		// Still awaiting payment.
		booking := setupTieredBooking(t, ctx, 72*time.Hour, false)

		lookup := lookupByToken(t, booking.Token)
		assert.Nil(t, lookup.RefundPreview, "no preview on a booking that cannot be cancelled")

		status, _ := cancelByToken(t, booking.ID, booking.Token)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Empty(t, payments.RefundsFor(*booking.PaymentIntentID))

		stored, err := bookingRepo.GetByID(ctx, booking.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.BookingStatusAwaitingPayment, stored.Status)
		assert.Nil(t, stored.CancelledBy)
	})
}
