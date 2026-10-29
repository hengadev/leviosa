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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// make test-func TEST_NAME=TestPractitionerCancellation TEST_PATH=test/integration/booking/booking/practitioner_cancellation_test.go

// Tiers under which a client cancellation refunds nothing, whatever the
// notice: a practitioner cancellation must refund everything all the same.
const practitionerTestTiers = `[{"minHoursBefore":0,"refundPercentage":0}]`

// cancelAsPractitioner posts the practitioner cancellation with the given
// session, no body (no reason is collected, ADR-0019).
func cancelAsPractitioner(t *testing.T, client *http.Client, bookingID uuid.UUID, token string) (int, domain.BookingResponse) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/partners/bookings/%s/cancel", testServerURL, bookingID), nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: token})
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var result domain.BookingResponse
	if resp.StatusCode == http.StatusOK {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	}
	return resp.StatusCode, result
}

func agendaEntry(t *testing.T, client *http.Client, f slotFixture, bookingID uuid.UUID) domain.PartnerBookingResponse {
	t.Helper()
	var bookings []domain.PartnerBookingResponse
	require.NoError(t, json.Unmarshal(fetchPartnerAgenda(t, client, f.partnerUserID, f.partnerToken), &bookings))
	for _, b := range bookings {
		if b.ID == bookingID {
			return b
		}
	}
	t.Fatalf("booking %s not in the partner agenda", bookingID)
	return domain.PartnerBookingResponse{}
}

func TestPractitionerCancellation(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}
	payments := mockPayments(t)

	t.Run("refunds everything paid online, frees the slot and emails the client once", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)
		day := f.day.Format(time.DateOnly)

		_, err := testPool.Exec(ctx, `UPDATE catalog.products SET deposit = $2, cancellation_tiers = $3 WHERE id = $1`,
			f.productID, fmt.Sprintf(`{"type":"fixed","value":%d}`, refundTestDepositCents), practitionerTestTiers)
		require.NoError(t, err)

		booking, err := service.CreateBooking(ctx, f.availabilityID, nil, f.productID, f.at(10, 15),
			"Jane", "Doe", "jane.practitioner@example.com", "+33612345678", domain.BookingLocaleEN)
		require.NoError(t, err)
		event := paymentSucceededEvent(t, ctx, booking.ID)
		event.Amount = refundTestDepositCents
		require.NoError(t, paymentService.ConfirmPaymentIntent(ctx, event.PaymentIntentID))
		require.NoError(t, service.HandlePaymentWebhook(ctx, event))

		// Booked: the slot is no longer offered; the agenda shows it upcoming,
		// with the full refund a cancellation would give.
		status, body := getFreeSlots(t, client, f.partnerID, f.productID, day, day)
		require.Equal(t, http.StatusOK, status)
		assertSameInstants(t, []time.Time{f.at(9, 0), f.at(11, 30)}, slotStarts(body))
		entry := agendaEntry(t, client, f, booking.ID)
		assert.Equal(t, domain.BookingStatusConfirmed, entry.Status)
		assert.Equal(t, refundTestDepositCents, entry.PractitionerRefundCents)

		// Another partner cannot cancel it.
		otherToken, _ := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "other-partner@example.com", f.roomID, testPool, authCtx.Redis, crypto)
		status, _ = cancelAsPractitioner(t, client, booking.ID, otherToken)
		assert.Equal(t, http.StatusForbidden, status)
		assert.Empty(t, payments.RefundsFor(*booking.PaymentIntentID), "a refused cancellation refunds nothing")
		assert.Equal(t, domain.BookingStatusConfirmed, storedBooking(t, ctx, booking.ID).Status)

		notificationEmailSpy.Reset()
		smsBefore := notificationSMSSpy.Count()
		stop := notificationSpy.ForwardConfirmationsTo(realNotificationAdapter)
		defer stop()

		status, result := cancelAsPractitioner(t, client, booking.ID, f.partnerToken)
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, domain.BookingStatusCancelled, result.Status)
		assert.Equal(t, domain.CancelledByPractitioner, result.CancelledBy)
		require.NotNil(t, result.RefundedAmountCents)
		assert.Equal(t, refundTestDepositCents, *result.RefundedAmountCents)
		assert.Equal(t, domain.PaymentStatusRefunded, result.PaymentStatus)

		// The payment port refunded the whole amount paid online, although
		// the tiers would have refunded nothing to the client.
		refunds := payments.RefundsFor(*booking.PaymentIntentID)
		require.Len(t, refunds, 1)
		assert.Equal(t, refundTestDepositCents, refunds[0].Amount)

		stored, err := bookingRepo.GetByID(ctx, booking.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.BookingStatusCancelled, stored.Status)
		assert.Equal(t, domain.PaymentStatusRefunded, stored.PaymentStatus)
		require.NotNil(t, stored.CancelledBy)
		assert.Equal(t, "practitioner", *stored.CancelledBy)
		require.NotNil(t, stored.RefundedAmountCents)
		assert.Equal(t, refundTestDepositCents, *stored.RefundedAmountCents)

		// It leaves the upcoming bookings, and its slot is offered again.
		entry = agendaEntry(t, client, f, booking.ID)
		assert.Equal(t, domain.BookingStatusCancelled, entry.Status)
		assert.Zero(t, entry.PractitionerRefundCents)
		status, body = getFreeSlots(t, client, f.partnerID, f.productID, day, day)
		require.Equal(t, http.StatusOK, status)
		assertSameInstants(t, []time.Time{f.at(9, 0), f.at(10, 15), f.at(11, 30)}, slotStarts(body))

		require.Eventually(t, func() bool {
			clients, _ := notificationEmailSpy.CancellationSnapshot()
			return len(clients) >= 1
		}, 5*time.Second, 20*time.Millisecond, "expected the client cancellation email")

		// A second cancel is refused and refunds nothing more.
		status, _ = cancelAsPractitioner(t, client, booking.ID, f.partnerToken)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Len(t, payments.RefundsFor(*booking.PaymentIntentID), 1, "no second refund")

		time.Sleep(200 * time.Millisecond) // give a duplicate email a chance to show up
		clients, partners := notificationEmailSpy.CancellationSnapshot()
		require.Len(t, clients, 1, "exactly one client cancellation email")
		client0 := clients[0]
		assert.Equal(t, "jane.practitioner@example.com", client0.ToEmail)
		assert.Equal(t, notificationDomain.LocaleEN, client0.Locale)
		assert.Equal(t, notificationDomain.CancelledByPractitioner, client0.CancelledBy)
		assert.Equal(t, refundTestDepositCents, client0.AmountPaidOnlineCents)
		assert.Equal(t, refundTestDepositCents, client0.RefundedAmountCents)

		// The partner gets a short confirmation of their own cancellation.
		require.Len(t, partners, 1)
		assert.Equal(t, notificationDomain.CancelledByPractitioner, partners[0].CancelledBy)

		assert.Equal(t, smsBefore, notificationSMSSpy.Count(), "booking notifications send no SMS")
	})

	t.Run("an unknown booking is not found", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)
		status, _ := cancelAsPractitioner(t, client, uuid.New(), f.partnerToken)
		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("a booking awaiting payment cannot be cancelled this way", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)
		booking, err := service.CreateBooking(ctx, f.availabilityID, nil, f.productID, f.at(9, 0),
			"Paul", "Martin", "paul@example.com", "0611112222", domain.BookingLocaleFR)
		require.NoError(t, err)

		status, _ := cancelAsPractitioner(t, client, booking.ID, f.partnerToken)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, domain.BookingStatusAwaitingPayment, storedBooking(t, ctx, booking.ID).Status)
	})
}
