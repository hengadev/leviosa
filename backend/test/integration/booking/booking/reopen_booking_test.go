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

// make test-func TEST_NAME=TestReopenBooking TEST_PATH=test/integration/booking/booking/reopen_booking_test.go

// setupClosedBooking creates a confirmed booking and returns its ID and the
// partner token, mirroring mark_no_show_test.go's setupBooking.
func setupClosedBooking(t *testing.T) (uuid.UUID, string) {
	t.Helper()
	ctx := context.Background()

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

	partnerToken, partnerID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "partner-reopen@test.com", room.ID, testPool, authCtx.Redis, crypto)
	clientToken, clientID := tsetup.SetupStandardUser(t, ctx, "client-reopen@test.com", room.ID, testPool, authCtx.Redis, crypto)

	startTime := time.Now().Add(48 * time.Hour).Truncate(time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	availability := tavail.NewTestAvailability(t, partnerID, room.ID, startTime, endTime, 5000)
	availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
	require.NoError(t, err)
	tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)

	// Create a booking via HTTP (first test product, 60-minute massage), then
	// have Stripe confirm it (ADR-0018).
	requestBody := map[string]interface{}{
		"availability_id": availability.ID.String(),
		"client_id":       clientID,
		"product_id":      products[0].ID.String(),
		"slot_start_time": startTime.Format(time.RFC3339),
	}
	bodyBytes, err := json.Marshal(requestBody)
	require.NoError(t, err)

	reqCreate, err := http.NewRequest("POST", testServerURL+"/bookings", bytes.NewReader(bodyBytes))
	require.NoError(t, err)
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: clientToken})

	client := &http.Client{Timeout: 10 * time.Second}
	respCreate, err := client.Do(reqCreate)
	require.NoError(t, err)
	defer respCreate.Body.Close()
	require.Equal(t, http.StatusCreated, respCreate.StatusCode)

	var createdBooking domain.BookingResponse
	require.NoError(t, json.NewDecoder(respCreate.Body).Decode(&createdBooking))
	confirmBookingPayment(t, ctx, createdBooking.ID)

	return createdBooking.ID, partnerToken
}

func TestReopenBooking(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("should reopen a completed booking the same day", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, partnerToken := setupClosedBooking(t)

		reqComplete, _ := http.NewRequest("POST", testServerURL+"/bookings/"+bookingID.String()+"/complete", nil)
		reqComplete.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: partnerToken})
		respComplete, err := client.Do(reqComplete)
		require.NoError(t, err)
		respComplete.Body.Close()
		require.Equal(t, http.StatusOK, respComplete.StatusCode)

		req, err := http.NewRequest("POST", testServerURL+"/bookings/"+bookingID.String()+"/reopen", nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: partnerToken})

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		dbBooking, err := bookingRepo.GetByID(ctx, bookingID)
		require.NoError(t, err)
		booking, err := domain.DecryptBookingEncx(ctx, crypto, dbBooking)
		require.NoError(t, err)
		assert.Equal(t, domain.BookingStatusConfirmed, booking.Status)
		assert.Nil(t, booking.CompletedAt)
	})

	t.Run("should reopen a no-show booking the same day", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, partnerToken := setupClosedBooking(t)

		reqNoShow, _ := http.NewRequest("POST", testServerURL+"/bookings/"+bookingID.String()+"/no-show", nil)
		reqNoShow.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: partnerToken})
		respNoShow, err := client.Do(reqNoShow)
		require.NoError(t, err)
		respNoShow.Body.Close()
		require.Equal(t, http.StatusOK, respNoShow.StatusCode)

		req, err := http.NewRequest("POST", testServerURL+"/bookings/"+bookingID.String()+"/reopen", nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: partnerToken})

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		dbBooking, err := bookingRepo.GetByID(ctx, bookingID)
		require.NoError(t, err)
		booking, err := domain.DecryptBookingEncx(ctx, crypto, dbBooking)
		require.NoError(t, err)
		assert.Equal(t, domain.BookingStatusConfirmed, booking.Status)
		assert.Nil(t, booking.NoShowAt)
	})

	t.Run("should return 409 when the day the booking was closed has passed", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, partnerToken := setupClosedBooking(t)

		reqComplete, _ := http.NewRequest("POST", testServerURL+"/bookings/"+bookingID.String()+"/complete", nil)
		reqComplete.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: partnerToken})
		respComplete, err := client.Do(reqComplete)
		require.NoError(t, err)
		respComplete.Body.Close()
		require.Equal(t, http.StatusOK, respComplete.StatusCode)

		// Shift the completion back one day to simulate a next-day reopen.
		_, err = testPool.Exec(ctx,
			"UPDATE booking.bookings SET completed_at = completed_at - interval '1 day' WHERE id = $1",
			bookingID)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServerURL+"/bookings/"+bookingID.String()+"/reopen", nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: partnerToken})

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("should return 404 for non-existent booking", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		partnerToken := tu.SetupPartnerUser(t, ctx, authCtx)

		req, err := http.NewRequest("POST", testServerURL+"/bookings/"+uuid.New().String()+"/reopen", nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: partnerToken})

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
