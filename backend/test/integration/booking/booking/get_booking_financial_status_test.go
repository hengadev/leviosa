package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
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

// Issue 011: the client booking detail page needs a plain-language
// cancellation deadline and refund status. These tests exercise the real
// GetBooking HTTP endpoint (real Postgres + encx, mocked Stripe per this
// package's existing convention) to confirm BookingResponse actually carries
// cancellation_deadline/cancellation_hours/refund_amount_cents/refunded_at
// for the states the client UI renders.
//
// make test-func TEST_NAME=TestGetBookingFinancialStatusOverHTTP TEST_PATH=test/integration/booking/booking/get_booking_financial_status_test.go

// setupBookingAt mirrors the setupBooking helper in get_booking_test.go, but
// lets the caller control how far the slot is from "now" so both sides of
// the cancellation deadline can be exercised. Test products (see
// test/helpers/catalog) all carry a hardcoded 24-hour cancellation window.
func setupBookingAt(t *testing.T, ctx context.Context, client *http.Client, slotOffset time.Duration) (bookingID uuid.UUID, clientToken string, adminToken string) {
	t.Helper()

	tbooking.ClearBookingsTable(t, ctx, testPool)
	tavail.ClearAvailabilityTable(t, ctx, testPool)
	tr.ClearRoomsTable(t, ctx, testPool)
	tb.ClearBuildingsTable(t, ctx, testPool)
	tcatalog.ClearProductsTable(t, ctx, testPool)

	categoryID := tcatalog.CreateTestCategory(t, ctx, testPool)
	products := tcatalog.CreateDefaultTestProducts(t, ctx, testPool, categoryID)

	building := tb.NewTestBuilding(t)
	buildingEncx, err := domain.ProcessBuildingEncx(ctx, crypto, building)
	require.NoError(t, err)
	require.NoError(t, tb.InsertBuildingEncx(t, ctx, testPool, buildingEncx))

	room := tr.NewTestRoomWithBuilding(t, building.ID)
	roomEncx, err := domain.ProcessRoomEncx(ctx, crypto, room)
	require.NoError(t, err)
	require.NoError(t, tr.InsertRoomEncx(t, ctx, testPool, roomEncx))

	adminToken, _ = tsetup.SetupAdminWithAllocation(t, ctx, room.ID, testPool, authCtx.Redis, crypto)
	_, partnerID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "partner-fin@test.com", room.ID, testPool, authCtx.Redis, crypto)
	clientToken, clientID := tsetup.SetupStandardUser(t, ctx, "client-fin@test.com", room.ID, testPool, authCtx.Redis, crypto)

	slotStartTime := time.Now().Add(slotOffset).Truncate(10 * time.Minute)
	availability := tavail.NewTestAvailability(t, partnerID, room.ID, slotStartTime.Add(-10*time.Minute), slotStartTime.Add(2*time.Hour), 5000)
	availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
	require.NoError(t, err)
	tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)

	productID := products[0].ID // "60-Minute Massage", CancellationHours: 24, PriceCents: 8000
	requestBody := map[string]interface{}{
		"availability_id": availability.ID.String(),
		"client_id":       clientID,
		"product_id":      productID.String(),
		"slot_start_time": slotStartTime.Format(time.RFC3339),
	}
	bodyBytes, err := json.Marshal(requestBody)
	require.NoError(t, err)

	reqCreate, err := http.NewRequest("POST", testServerURL+"/bookings", bytes.NewReader(bodyBytes))
	require.NoError(t, err)
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: clientToken})

	respCreate, err := client.Do(reqCreate)
	require.NoError(t, err)
	defer respCreate.Body.Close()
	require.Equal(t, http.StatusCreated, respCreate.StatusCode)

	var createdBooking domain.BookingResponse
	require.NoError(t, json.NewDecoder(respCreate.Body).Decode(&createdBooking))

	return createdBooking.ID, clientToken, adminToken
}

func fetchBooking(t *testing.T, client *http.Client, bookingID uuid.UUID, token string) domain.BookingResponse {
	t.Helper()

	req, err := http.NewRequest("GET", testServerURL+"/bookings/"+bookingID.String(), nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: token})

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var booking domain.BookingResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&booking))
	return booking
}

func TestGetBookingFinancialStatusOverHTTP(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("shows a free-cancellation deadline while still inside the window", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		// Slot 48h out with a 24h cancellation policy: the deadline sits
		// ~24h from now, still in the future.
		bookingID, clientToken, _ := setupBookingAt(t, ctx, client, 48*time.Hour)

		booking := fetchBooking(t, client, bookingID, clientToken)

		require.NotNil(t, booking.CancellationDeadline, "expected a cancellation deadline for a confirmed booking with a cancellation policy")
		require.NotNil(t, booking.CancellationHours)
		assert.Equal(t, 24, *booking.CancellationHours)
		assert.True(t, booking.CancellationDeadline.After(time.Now()), "deadline should still be in the future")
		assert.WithinDuration(t, booking.SlotStartTime.Add(-24*time.Hour), *booking.CancellationDeadline, time.Second)
		assert.Nil(t, booking.RefundAmountCents)
		assert.Nil(t, booking.RefundedAt)
	})

	t.Run("shows a past cancellation deadline once the window has closed", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		// Slot 2h out with a 24h cancellation policy: the deadline sits
		// ~22h in the past, even though the booking is still
		// confirmed/upcoming — this is the "fee applies" copy case.
		bookingID, clientToken, _ := setupBookingAt(t, ctx, client, 2*time.Hour)

		booking := fetchBooking(t, client, bookingID, clientToken)

		require.NotNil(t, booking.CancellationDeadline)
		assert.True(t, booking.CancellationDeadline.Before(time.Now()), "deadline should already be in the past")
	})

	t.Run("shows refund amount and date once the booking has been refunded", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, clientToken, adminToken := setupBookingAt(t, ctx, client, 72*time.Hour)

		// Drive the booking to "paid" the same way get_partner_earnings_test
		// does: flip the mock payment intent to succeeded, then mark the
		// booking paid directly (no real Stripe round trip in this suite).
		bookingEncx, err := bookingRepo.GetByID(ctx, bookingID)
		require.NoError(t, err)
		paidBooking, err := domain.DecryptBookingEncx(ctx, crypto, bookingEncx)
		require.NoError(t, err)
		require.NotNil(t, paidBooking.PaymentIntentID)

		mockPayments, ok := paymentService.(*MockPaymentService)
		require.True(t, ok, "test payment service must be the mock for this assertion to be meaningful")
		require.NoError(t, mockPayments.ConfirmPaymentIntent(ctx, *paidBooking.PaymentIntentID))

		require.NoError(t, paidBooking.MarkPaymentPaid())
		paidBookingEncx, err := domain.ProcessBookingEncx(ctx, crypto, paidBooking)
		require.NoError(t, err)
		require.NoError(t, bookingRepo.Update(ctx, paidBookingEncx))

		// Refund via the real admin-facing HTTP endpoint (RefundBooking).
		reqRefund, err := http.NewRequest("POST", testServerURL+"/bookings/"+bookingID.String()+"/refund", nil)
		require.NoError(t, err)
		reqRefund.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: adminToken})
		respRefund, err := client.Do(reqRefund)
		require.NoError(t, err)
		defer respRefund.Body.Close()
		require.Equal(t, http.StatusOK, respRefund.StatusCode)

		booking := fetchBooking(t, client, bookingID, clientToken)

		require.NotNil(t, booking.RefundAmountCents)
		assert.Equal(t, booking.TotalPriceCents, *booking.RefundAmountCents)
		require.NotNil(t, booking.RefundedAt)
		assert.WithinDuration(t, time.Now(), *booking.RefundedAt, 10*time.Second)
	})
}
