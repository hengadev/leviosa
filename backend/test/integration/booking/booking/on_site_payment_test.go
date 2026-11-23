package booking

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// make test-func TEST_NAME=TestOnSitePayment TEST_PATH=test/integration/booking/booking/on_site_payment_test.go

const onSiteDepositCents = 2000

// startSessionWithDeposit moves the booking's session to have started half
// an hour ago and records that only a deposit was paid online. The slot
// times are encrypted, so they go through the repository.
func startSessionWithDeposit(t *testing.T, ctx context.Context, bookingID uuid.UUID) *domain.Booking {
	t.Helper()
	booking := storedOnSiteBooking(t, ctx, bookingID)
	booking.SlotStartTime = time.Now().Add(-30 * time.Minute)
	booking.SlotEndTime = booking.SlotStartTime.Add(time.Hour)
	bookingEncx, err := domain.ProcessBookingEncx(ctx, crypto, booking)
	require.NoError(t, err)
	require.NoError(t, bookingRepo.Update(ctx, bookingEncx))

	_, err = testPool.Exec(ctx,
		"UPDATE booking.bookings SET amount_paid_online_cents = $2 WHERE id = $1",
		bookingID, onSiteDepositCents)
	require.NoError(t, err)

	booking = storedOnSiteBooking(t, ctx, bookingID)
	require.Greater(t, booking.TotalPriceCents, onSiteDepositCents, "a deposit, not the full price")
	return booking
}

func storedOnSiteBooking(t *testing.T, ctx context.Context, bookingID uuid.UUID) *domain.Booking {
	t.Helper()
	bookingEncx, err := bookingRepo.GetByID(ctx, bookingID)
	require.NoError(t, err)
	booking, err := domain.DecryptBookingEncx(ctx, crypto, bookingEncx)
	require.NoError(t, err)
	return booking
}

func onSitePaymentRequest(t *testing.T, method string, bookingID uuid.UUID, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, testServerURL+"/bookings/"+bookingID.String()+"/on-site-payment", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: token})
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	require.NoError(t, err)
	return resp
}

func TestOnSitePayment(t *testing.T) {
	ctx := context.Background()

	t.Run("should record the amount due at the practice on a started session", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, partnerToken := setupClosedBooking(t)
		booking := startSessionWithDeposit(t, ctx, bookingID)
		due := booking.TotalPriceCents - onSiteDepositCents

		resp := onSitePaymentRequest(t, http.MethodPost, bookingID, partnerToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var body domain.BookingResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		assert.Equal(t, due, body.AmountDueOnSiteCents)
		require.NotNil(t, body.AmountPaidOnSiteCents)
		assert.Equal(t, due, *body.AmountPaidOnSiteCents)

		stored := storedOnSiteBooking(t, ctx, bookingID)
		require.NotNil(t, stored.AmountPaidOnSiteCents)
		assert.Equal(t, due, *stored.AmountPaidOnSiteCents)
		require.NotNil(t, stored.PaidOnSiteAt)
		assert.WithinDuration(t, time.Now(), *stored.PaidOnSiteAt, time.Minute)
		assert.Equal(t, domain.BookingStatusConfirmed, stored.Status)

		// The partner bookings and earnings carry the amounts.
		partnerBookings, err := service.GetPartnerBookingsEnriched(ctx, booking.PartnerID, ports.BookingFilter{})
		require.NoError(t, err)
		require.Len(t, partnerBookings, 1)
		assert.Equal(t, onSiteDepositCents, partnerBookings[0].AmountPaidOnlineCents)
		assert.Equal(t, due, partnerBookings[0].AmountDueOnSiteCents)
		require.NotNil(t, partnerBookings[0].AmountPaidOnSiteCents)
		assert.Equal(t, due, *partnerBookings[0].AmountPaidOnSiteCents)

		earnings, err := service.GetPartnerEarnings(ctx, booking.PartnerID)
		require.NoError(t, err)
		require.Len(t, earnings.Transactions, 1)
		assert.Equal(t, due, earnings.Transactions[0].AmountDueOnSiteCents)
		require.NotNil(t, earnings.Transactions[0].AmountPaidOnSiteCents)
		assert.Equal(t, due, *earnings.Transactions[0].AmountPaidOnSiteCents)
	})

	t.Run("should let an administrator record it", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, _ := setupClosedBooking(t)
		startSessionWithDeposit(t, ctx, bookingID)
		adminToken := tu.SetupAdminUser(t, ctx, authCtx)

		resp := onSitePaymentRequest(t, http.MethodPost, bookingID, adminToken)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("should return 409 when already recorded", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, partnerToken := setupClosedBooking(t)
		startSessionWithDeposit(t, ctx, bookingID)

		first := onSitePaymentRequest(t, http.MethodPost, bookingID, partnerToken)
		first.Body.Close()
		require.Equal(t, http.StatusOK, first.StatusCode)

		resp := onSitePaymentRequest(t, http.MethodPost, bookingID, partnerToken)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
		var body map[string]string
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		assert.Equal(t, domain.ErrOnSitePaymentAlreadyRecorded.Error(), body["error"])
	})

	t.Run("should return 409 before the session starts", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, partnerToken := setupClosedBooking(t)
		_, err := testPool.Exec(ctx,
			"UPDATE booking.bookings SET amount_paid_online_cents = $2 WHERE id = $1",
			bookingID, onSiteDepositCents)
		require.NoError(t, err)

		resp := onSitePaymentRequest(t, http.MethodPost, bookingID, partnerToken)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
		assert.Nil(t, storedOnSiteBooking(t, ctx, bookingID).AmountPaidOnSiteCents)
	})

	t.Run("should return 409 on a no-show, keeping its deposit", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, partnerToken := setupClosedBooking(t)
		startSessionWithDeposit(t, ctx, bookingID)
		_, err := testPool.Exec(ctx,
			"UPDATE booking.bookings SET status = 'no_show', no_show_at = NOW() WHERE id = $1", bookingID)
		require.NoError(t, err)

		resp := onSitePaymentRequest(t, http.MethodPost, bookingID, partnerToken)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		stored := storedOnSiteBooking(t, ctx, bookingID)
		assert.Nil(t, stored.AmountPaidOnSiteCents)
		require.NotNil(t, stored.AmountPaidOnlineCents)
		assert.Equal(t, onSiteDepositCents, *stored.AmountPaidOnlineCents)
	})

	t.Run("should return 404 for another practitioner's booking", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, _ := setupClosedBooking(t)
		startSessionWithDeposit(t, ctx, bookingID)
		otherToken := tu.SetupPartnerUser(t, ctx, authCtx)

		resp := onSitePaymentRequest(t, http.MethodPost, bookingID, otherToken)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Nil(t, storedOnSiteBooking(t, ctx, bookingID).AmountPaidOnSiteCents)
	})

	t.Run("should undo the payment the same day", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, partnerToken := setupClosedBooking(t)
		startSessionWithDeposit(t, ctx, bookingID)
		record := onSitePaymentRequest(t, http.MethodPost, bookingID, partnerToken)
		record.Body.Close()
		require.Equal(t, http.StatusOK, record.StatusCode)

		resp := onSitePaymentRequest(t, http.MethodDelete, bookingID, partnerToken)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		stored := storedOnSiteBooking(t, ctx, bookingID)
		assert.Nil(t, stored.AmountPaidOnSiteCents)
		assert.Nil(t, stored.PaidOnSiteAt)
	})

	t.Run("should return 409 when undoing after the day it was recorded", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, partnerToken := setupClosedBooking(t)
		startSessionWithDeposit(t, ctx, bookingID)
		record := onSitePaymentRequest(t, http.MethodPost, bookingID, partnerToken)
		record.Body.Close()
		require.Equal(t, http.StatusOK, record.StatusCode)

		// Shift the recording back one day to simulate a next-day undo.
		_, err := testPool.Exec(ctx,
			"UPDATE booking.bookings SET paid_on_site_at = paid_on_site_at - interval '1 day' WHERE id = $1",
			bookingID)
		require.NoError(t, err)

		resp := onSitePaymentRequest(t, http.MethodDelete, bookingID, partnerToken)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
		assert.NotNil(t, storedOnSiteBooking(t, ctx, bookingID).AmountPaidOnSiteCents)
	})

	t.Run("should return 409 when undoing a payment never recorded", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, partnerToken := setupClosedBooking(t)

		resp := onSitePaymentRequest(t, http.MethodDelete, bookingID, partnerToken)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})
}
