package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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

// ADR-0019: Leviosa stores no health data. The booking API neither accepts
// nor returns client_notes/partner_notes. The DB columns stay, but no write
// path reaches them and no response contains them.
//
// make test-func TEST_NAME=TestBookingNotesRejectedByAPI TEST_PATH=test/integration/booking/booking/notes_rejected_test.go

func TestBookingNotesRejectedByAPI(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	// Helper that creates a booking and returns its ID plus the client token.
	setupBooking := func(t *testing.T) (uuid.UUID, string) {
		t.Helper()

		// Clean state
		tbooking.ClearBookingsTable(t, ctx, testPool)
		tavail.ClearAvailabilityTable(t, ctx, testPool)
		tr.ClearRoomsTable(t, ctx, testPool)
		tb.ClearBuildingsTable(t, ctx, testPool)
		tcatalog.ClearProductsTable(t, ctx, testPool)

		// Setup test category and products
		categoryID := tcatalog.CreateTestCategory(t, ctx, testPool)
		products := tcatalog.CreateDefaultTestProducts(t, ctx, testPool, categoryID)

		// Setup test building and room
		building := tb.NewTestBuilding(t)
		buildingEncx, err := domain.ProcessBuildingEncx(ctx, crypto, building)
		require.NoError(t, err)
		err = tb.InsertBuildingEncx(t, ctx, testPool, buildingEncx)
		require.NoError(t, err)

		room := tr.NewTestRoomWithBuilding(t, building.ID)
		roomEncx, err := domain.ProcessRoomEncx(ctx, crypto, room)
		require.NoError(t, err)
		err = tr.InsertRoomEncx(t, ctx, testPool, roomEncx)
		require.NoError(t, err)

		// Setup authenticated partner and client
		_, partnerID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "partner@test.com", room.ID, testPool, authCtx.Redis, crypto)
		clientToken, clientID := tsetup.SetupStandardUser(t, ctx, "client@test.com", room.ID, testPool, authCtx.Redis, crypto)

		// Create availability
		startTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
		endTime := startTime.Add(2 * time.Hour)
		availability := tavail.NewTestAvailability(t, partnerID, room.ID, startTime, endTime, 5000)
		availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
		require.NoError(t, err)
		tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)
		require.NoError(t, err)

		// Create booking via HTTP using first test product (60-minute massage)
		productID := products[0].ID
		slotStartTime := startTime // the availability's first slot (slot engine grid)
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

		body, err := io.ReadAll(respCreate.Body)
		require.NoError(t, err)
		var createdBooking domain.BookingResponse
		require.NoError(t, json.Unmarshal(body, &createdBooking))

		// No notes keys in the create response
		var raw map[string]any
		require.NoError(t, json.Unmarshal(body, &raw))
		assert.NotContains(t, raw, "client_notes")
		assert.NotContains(t, raw, "partner_notes")

		return createdBooking.ID, clientToken
	}

	t.Run("should reject client_notes in the create request", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		_, clientToken := setupBooking(t)

		// A second availability is not needed: the request must be rejected on
		// decode, before any booking logic runs.
		requestBody := map[string]interface{}{
			"availability_id": uuid.New().String(),
			"client_id":       uuid.New().String(),
			"product_id":      uuid.New().String(),
			"slot_start_time": time.Now().Add(48 * time.Hour).Format(time.RFC3339),
			"client_notes":    "My back hurts",
		}
		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServerURL+"/bookings", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: clientToken})

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Unknown fields are rejected: the API does not accept notes writes
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return 404 for the removed notes endpoint", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, clientToken := setupBooking(t)

		req, err := http.NewRequest("PUT", testServerURL+"/bookings/"+bookingID.String()+"/notes", bytes.NewReader([]byte(`{"partner_notes":"private"}`)))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: clientToken})

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// The route is gone: writes are rejected
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Nothing landed in the legacy columns either
		dbBookingEncx, err := bookingRepo.GetByID(ctx, bookingID)
		require.NoError(t, err)
		dbBooking, err := domain.DecryptBookingEncx(ctx, crypto, dbBookingEncx)
		require.NoError(t, err)
		assert.Empty(t, dbBooking.ClientNotes)
		assert.Empty(t, dbBooking.PartnerNotes)
	})
}
