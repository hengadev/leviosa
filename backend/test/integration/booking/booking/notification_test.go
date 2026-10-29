package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookingConfirmationNotification(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	// Booking lifecycle (ADR-0018): create → awaiting_payment, nothing sent;
	// payment webhook → confirmed, booking-confirmed event emitted once, even
	// when Stripe delivers the webhook twice.
	t.Run("should send the booking confirmation once Stripe confirms the payment, never at creation", func(t *testing.T) {
		// Clean state
		tbooking.ClearBookingsTable(t, ctx, testPool)
		tavail.ClearAvailabilityTable(t, ctx, testPool)
		tr.ClearRoomsTable(t, ctx, testPool)
		tb.ClearBuildingsTable(t, ctx, testPool)
		tcatalog.ClearProductsTable(t, ctx, testPool)
		defer tu.ClearAuthData(t, ctx, authCtx)

		// Reset spy
		notificationSpy.BookingConfirmations = nil
		notificationSpy.PaymentConfirmations = nil

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
		_, partnerID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "partner-notify@example.com", room.ID, testPool, authCtx.Redis, crypto)
		clientToken, clientID := tsetup.SetupStandardUser(t, ctx, "client-notify@example.com", room.ID, testPool, authCtx.Redis, crypto)

		// Create test availability
		startTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
		endTime := startTime.Add(2 * time.Hour)
		availability := tavail.NewTestAvailability(t, partnerID, room.ID, startTime, endTime, 0)
		availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
		require.NoError(t, err)
		tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)

		// Prepare booking request
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

		// Parse response to get booking ID
		var booking domain.CreateBookingResponse
		err = json.NewDecoder(resp.Body).Decode(&booking)
		require.NoError(t, err)

		// The booking awaits payment, held for the default 15 minutes, with
		// the PaymentIntent's client secret for the Payment Element.
		assert.Equal(t, domain.BookingStatusAwaitingPayment, booking.Status)
		assert.NotEmpty(t, booking.ClientSecret)
		assert.Equal(t, booking.TotalPriceCents, booking.AmountDueOnlineCents, "no deposit: the full price is due online")
		require.NotNil(t, booking.HoldExpiresAt)
		assert.WithinDuration(t, time.Now().Add(15*time.Minute), *booking.HoldExpiresAt, time.Minute)
		assert.Empty(t, notificationSpy.BookingConfirmations, "nothing is sent while the booking awaits payment")

		// Stripe delivers payment_intent.succeeded twice.
		event := paymentSucceededEvent(t, ctx, booking.ID)
		require.NoError(t, service.HandlePaymentWebhook(ctx, event))
		require.NoError(t, service.HandlePaymentWebhook(ctx, event))

		bookingEncx, err := bookingRepo.GetByID(ctx, booking.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.BookingStatusConfirmed, bookingEncx.Status)
		assert.Equal(t, domain.PaymentStatusPaid, bookingEncx.PaymentStatus)

		// Assert notification was sent once
		require.Len(t, notificationSpy.BookingConfirmations, 1, "expected exactly one booking confirmation notification")

		notification := notificationSpy.BookingConfirmations[0]
		assert.Equal(t, booking.ID, notification.BookingID, "notification should reference the created booking")
		assert.Equal(t, clientID, notification.ClientID, "notification should reference the client")
		assert.Equal(t, partnerID, notification.PartnerID, "notification should reference the partner")
		assert.Equal(t, productID, notification.ProductID, "notification should reference the product")
		assert.Equal(t, room.ID, notification.RoomID, "notification should reference the room")
		assert.Equal(t, slotStartTime.Unix(), notification.SlotStartTime.Unix(), "notification should have correct slot start time")
	})
}

func TestBookingCancellationNotification(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("should send cancellation notifications to both client and partner when booking is cancelled", func(t *testing.T) {
		// Clean state
		tbooking.ClearBookingsTable(t, ctx, testPool)
		tavail.ClearAvailabilityTable(t, ctx, testPool)
		tr.ClearRoomsTable(t, ctx, testPool)
		tb.ClearBuildingsTable(t, ctx, testPool)
		tcatalog.ClearProductsTable(t, ctx, testPool)
		defer tu.ClearAuthData(t, ctx, authCtx)

		// Reset spy
		notificationSpy.BookingConfirmations = nil
		notificationSpy.BookingCancellations = nil
		notificationSpy.PaymentConfirmations = nil

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
		_, partnerID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "partner-cancel@example.com", room.ID, testPool, authCtx.Redis, crypto)
		clientToken, clientID := tsetup.SetupStandardUser(t, ctx, "client-cancel@example.com", room.ID, testPool, authCtx.Redis, crypto)

		// Create test availability (far enough in the future to pass cancellation policy)
		startTime := time.Now().Add(72 * time.Hour).Truncate(time.Hour)
		endTime := startTime.Add(2 * time.Hour)
		availability := tavail.NewTestAvailability(t, partnerID, room.ID, startTime, endTime, 0)
		availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
		require.NoError(t, err)
		tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)

		// Create booking
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

		req, err := http.NewRequest("POST", testServerURL+"/bookings", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: clientToken})

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var createdBooking domain.BookingResponse
		err = json.NewDecoder(resp.Body).Decode(&createdBooking)
		require.NoError(t, err)

		// Nothing is sent at creation; the payment confirms the booking.
		require.Empty(t, notificationSpy.BookingConfirmations, "nothing is sent while the booking awaits payment")
		confirmBookingPayment(t, ctx, createdBooking.ID)
		require.Len(t, notificationSpy.BookingConfirmations, 1, "expected booking confirmation once paid")

		// Reset spy to isolate cancellation notifications
		notificationSpy.BookingConfirmations = nil
		notificationSpy.BookingCancellations = nil

		// Cancel the booking
		cancelRequest := domain.CancelBookingRequest{
			Reason: "Client needs to reschedule",
		}
		cancelBytes, err := json.Marshal(cancelRequest)
		require.NoError(t, err)

		cancelReq, err := http.NewRequest("POST", testServerURL+"/bookings/"+createdBooking.ID.String()+"/cancel", bytes.NewReader(cancelBytes))
		require.NoError(t, err)
		cancelReq.Header.Set("Content-Type", "application/json")
		cancelReq.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: clientToken})

		cancelResp, err := client.Do(cancelReq)
		require.NoError(t, err)
		defer cancelResp.Body.Close()
		assert.Equal(t, http.StatusOK, cancelResp.StatusCode)

		// Assert cancellation notification was sent
		require.Len(t, notificationSpy.BookingCancellations, 1, "expected exactly one booking cancellation notification call")

		notification := notificationSpy.BookingCancellations[0]
		assert.Equal(t, createdBooking.ID, notification.BookingID, "notification should reference the cancelled booking")
		assert.Equal(t, clientID, notification.ClientID, "notification should reference the client")
		assert.Equal(t, partnerID, notification.PartnerID, "notification should reference the partner")
		assert.Equal(t, productID, notification.ProductID, "notification should reference the product")
		assert.Equal(t, room.ID, notification.RoomID, "notification should reference the room")
		assert.Equal(t, "client", notification.CancelledBy, "notification should say the client cancelled")
		assert.NotNil(t, notification.CancelledAt, "notification should include cancelled_at timestamp")
	})
}

func TestPaymentConfirmationNotification(t *testing.T) {
	ctx := context.Background()
	httpClient := &http.Client{Timeout: 10 * time.Second}

	t.Run("should confirm the booking once when the payment webhook is delivered concurrently", func(t *testing.T) {
		// Clean state
		tbooking.ClearBookingsTable(t, ctx, testPool)
		tavail.ClearAvailabilityTable(t, ctx, testPool)
		tr.ClearRoomsTable(t, ctx, testPool)
		tb.ClearBuildingsTable(t, ctx, testPool)
		tcatalog.ClearProductsTable(t, ctx, testPool)
		defer tu.ClearAuthData(t, ctx, authCtx)

		// Reset spy
		notificationSpy.BookingConfirmations = nil
		notificationSpy.PaymentConfirmations = nil
		notificationSpy.PaymentFailed = nil

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
		_, partnerID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "partner-payment@example.com", room.ID, testPool, authCtx.Redis, crypto)
		clientToken, clientID := tsetup.SetupStandardUser(t, ctx, "client-payment@example.com", room.ID, testPool, authCtx.Redis, crypto)

		// Create test availability
		startTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
		endTime := startTime.Add(2 * time.Hour)
		availability := tavail.NewTestAvailability(t, partnerID, room.ID, startTime, endTime, 0)
		availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
		require.NoError(t, err)
		tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)

		// Prepare booking request
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

		req, err := http.NewRequest("POST", testServerURL+"/bookings", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: clientToken})

		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var booking domain.BookingResponse
		err = json.NewDecoder(resp.Body).Decode(&booking)
		require.NoError(t, err)

		require.Equal(t, domain.BookingStatusAwaitingPayment, booking.Status)

		// Concurrent deliveries of the same payment_intent.succeeded event:
		// the conditional confirmation lets exactly one of them through.
		event := paymentSucceededEvent(t, ctx, booking.ID)
		var wg sync.WaitGroup
		errsCh := make(chan error, 4)
		for range 4 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				errsCh <- service.HandlePaymentWebhook(ctx, event)
			}()
		}
		wg.Wait()
		close(errsCh)
		for err := range errsCh {
			require.NoError(t, err)
		}

		// The confirmation carries the payment: one booking confirmation, no
		// separate payment confirmation.
		require.Len(t, notificationSpy.BookingConfirmations, 1, "expected one booking confirmation after payment success")
		assert.Empty(t, notificationSpy.PaymentConfirmations)

		confirmation := notificationSpy.BookingConfirmations[0]
		assert.Equal(t, booking.ID, confirmation.BookingID, "the confirmation should reference the booking")
		assert.Equal(t, clientID, confirmation.ClientID, "the confirmation should reference the client")
	})
}
