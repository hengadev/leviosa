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

// make test-func TEST_NAME=TestCreateBooking TEST_PATH=test/integration/booking/booking/create_booking_test.go

func TestCreateBooking(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("should successfully create a booking", func(t *testing.T) {
		// Clean state
		tbooking.ClearBookingsTable(t, ctx, testPool)
		tavail.ClearAvailabilityTable(t, ctx, testPool)
		tr.ClearRoomsTable(t, ctx, testPool)
		tb.ClearBuildingsTable(t, ctx, testPool)
		tcatalog.ClearProductsTable(t, ctx, testPool)
		defer tu.ClearAuthData(t, ctx, authCtx)

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

		// Setup authenticated users
		_, partnerID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "partner@example.com", room.ID, testPool, authCtx.Redis, crypto)
		clientToken, clientID := tsetup.SetupStandardUser(t, ctx, "client@example.com", room.ID, testPool, authCtx.Redis, crypto)

		// Create test availability
		startTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
		endTime := startTime.Add(2 * time.Hour)
		availability := tavail.NewTestAvailability(t, partnerID, room.ID, startTime, endTime, 0)
		availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
		require.NoError(t, err)
		tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)

		// Prepare booking request with product ID and slot time using first test product (60-minute massage)
		productID := products[0].ID
		slotStartTime := startTime // the availability's first slot (slot engine grid)
		requestBody := map[string]interface{}{
			"availability_id": availability.ID.String(),
			"client_id":       clientID,
			"product_id":      productID.String(),
			"slot_start_time": slotStartTime.Format(time.RFC3339),
			"locale":          "en",
		}
		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServerURL+"/bookings", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: clientToken})

		// Execute request
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Assert HTTP response
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Parse response
		var booking domain.CreateBookingResponse
		err = json.NewDecoder(resp.Body).Decode(&booking)
		require.NoError(t, err)

		// Assert booking details
		assert.NotEqual(t, uuid.Nil, booking.ID)
		assert.Equal(t, availability.ID, booking.AvailabilityID)
		assert.Equal(t, &clientID, booking.ClientID)
		assert.Equal(t, partnerID, booking.PartnerID)
		assert.Equal(t, room.ID, booking.RoomID)
		assert.Equal(t, productID, booking.ProductID)
		assert.Equal(t, slotStartTime.Unix(), booking.SlotStartTime.Unix())
		assert.Equal(t, products[0].PriceCents, booking.TotalPriceCents) // the product's active one-time price
		assert.Equal(t, "EUR", booking.Currency)
		// A new booking awaits payment, held for the payment hold, and carries
		// what the page needs to collect the payment (ADR-0018).
		assert.Equal(t, domain.BookingStatusAwaitingPayment, booking.Status)
		assert.Equal(t, domain.PaymentStatusPending, booking.PaymentStatus)
		assert.NotEmpty(t, booking.ClientSecret, "the PaymentIntent's client secret")
		assert.NotNil(t, booking.PaymentIntentID)
		assert.Equal(t, products[0].PriceCents, booking.AmountDueOnlineCents, "no deposit: the full price is due online")
		require.NotNil(t, booking.HoldExpiresAt)
		assert.WithinDuration(t, time.Now().Add(15*time.Minute), *booking.HoldExpiresAt, time.Minute)
		assert.Equal(t, "en", booking.Locale)

		// Verify booking exists in database, with its hold and locale
		savedBooking, err := bookingRepo.GetByID(ctx, booking.ID)
		require.NoError(t, err)
		assert.Equal(t, booking.ID, savedBooking.ID)
		assert.Equal(t, domain.BookingStatusAwaitingPayment, savedBooking.Status)
		require.NotNil(t, savedBooking.HoldExpiresAt)
		require.NotNil(t, savedBooking.Locale)
		assert.Equal(t, "en", *savedBooking.Locale)
	})

	// Archiving a product keeps it for the bookings that refer to it but
	// stops offering it; a draft is not offered yet.
	for _, status := range []string{"archived", "draft"} {
		t.Run("should return 409 for a "+status+" product", func(t *testing.T) {
			tbooking.ClearBookingsTable(t, ctx, testPool)
			tavail.ClearAvailabilityTable(t, ctx, testPool)
			tr.ClearRoomsTable(t, ctx, testPool)
			tb.ClearBuildingsTable(t, ctx, testPool)
			tcatalog.ClearProductsTable(t, ctx, testPool)
			defer tu.ClearAuthData(t, ctx, authCtx)

			categoryID := tcatalog.CreateTestCategory(t, ctx, testPool)
			products := tcatalog.CreateDefaultTestProducts(t, ctx, testPool, categoryID)
			productID := products[0].ID
			_, err := testPool.Exec(ctx, `UPDATE catalog.products SET status = $2 WHERE id = $1`, productID, status)
			require.NoError(t, err)

			building := tb.NewTestBuilding(t)
			buildingEncx, err := domain.ProcessBuildingEncx(ctx, crypto, building)
			require.NoError(t, err)
			require.NoError(t, tb.InsertBuildingEncx(t, ctx, testPool, buildingEncx))
			room := tr.NewTestRoomWithBuilding(t, building.ID)
			roomEncx, err := domain.ProcessRoomEncx(ctx, crypto, room)
			require.NoError(t, err)
			require.NoError(t, tr.InsertRoomEncx(t, ctx, testPool, roomEncx))

			_, partnerID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "partner@example.com", room.ID, testPool, authCtx.Redis, crypto)
			clientToken, clientID := tsetup.SetupStandardUser(t, ctx, "client@example.com", room.ID, testPool, authCtx.Redis, crypto)

			startTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
			availability := tavail.NewTestAvailability(t, partnerID, room.ID, startTime, startTime.Add(2*time.Hour), 0)
			availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
			require.NoError(t, err)
			tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)

			bodyBytes, err := json.Marshal(map[string]interface{}{
				"availability_id": availability.ID.String(),
				"client_id":       clientID,
				"product_id":      productID.String(),
				"slot_start_time": startTime.Format(time.RFC3339),
			})
			require.NoError(t, err)
			req, err := http.NewRequest("POST", testServerURL+"/bookings", bytes.NewReader(bodyBytes))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: clientToken})

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusConflict, resp.StatusCode)

			var count int
			require.NoError(t, testPool.QueryRow(ctx, `SELECT count(*) FROM booking.bookings`).Scan(&count))
			assert.Zero(t, count, "no booking is created")
		})
	}

	t.Run("should return 400 for invalid availability", func(t *testing.T) {
		// Clean state
		tbooking.ClearBookingsTable(t, ctx, testPool)
		tavail.ClearAvailabilityTable(t, ctx, testPool)
		defer tu.ClearAuthData(t, ctx, authCtx)

		// Setup standard user for authentication
		userToken := tu.SetupStandardUser(t, ctx, authCtx)

		// Try to create booking with non-existent availability
		nonExistentAvailID := uuid.New()
		productID := uuid.New()
		slotStartTime := time.Now().Add(24 * time.Hour).Truncate(10 * time.Minute)
		requestBody := map[string]interface{}{
			"availability_id": nonExistentAvailID.String(),
			"client_id":       uuid.New(),
			"product_id":      productID.String(),
			"slot_start_time": slotStartTime.Format(time.RFC3339),
		}
		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServerURL+"/bookings", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: userToken})

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Assert 400 Bad Request
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return 400 for an unknown locale", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		userToken := tu.SetupStandardUser(t, ctx, authCtx)
		bodyBytes, err := json.Marshal(map[string]interface{}{
			"availability_id":  uuid.New().String(),
			"product_id":       uuid.New().String(),
			"slot_start_time":  time.Now().Add(24 * time.Hour).Truncate(time.Hour).Format(time.RFC3339),
			"guest_first_name": "Jean",
			"guest_last_name":  "Dupont",
			"guest_email":      "jean@example.com",
			"locale":           "de",
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServerURL+"/bookings", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: userToken})

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return 415 for missing content-type header", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		userToken := tu.SetupStandardUser(t, ctx, authCtx)

		req, err := http.NewRequest("POST", testServerURL+"/bookings", bytes.NewReader([]byte("{}")))
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: userToken})
		// Intentionally not setting Content-Type

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Assert 415 Unsupported Media Type
		assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
	})
}
