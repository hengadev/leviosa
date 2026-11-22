package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	bookingPorts "github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	settingsContracts "github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	settingsDomain "github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
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

// make test-func TEST_NAME=TestLookupBooking TEST_PATH=test/integration/booking/booking/lookup_booking_test.go

const (
	guestEmail = "alice@example.com"
	guestPhone = "0612345678"
)

// setupGuestBookingForLookup creates all infra and returns a guest booking with token.
func setupGuestBookingForLookup(t *testing.T) (bookingID uuid.UUID, token string) {
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
	err = tb.InsertBuildingEncx(t, ctx, testPool, buildingEncx)
	require.NoError(t, err)

	room := tr.NewTestRoomWithBuilding(t, building.ID)
	roomEncx, err := domain.ProcessRoomEncx(ctx, crypto, room)
	require.NoError(t, err)
	err = tr.InsertRoomEncx(t, ctx, testPool, roomEncx)
	require.NoError(t, err)

	startTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	_, partnerID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "partner@lookup-test.com", room.ID, testPool, authCtx.Redis, crypto)

	availability := tavail.NewTestAvailability(t, partnerID, room.ID, startTime, endTime, 5000)
	availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
	require.NoError(t, err)
	tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)

	productID := products[0].ID
	slotStartTime := startTime // the availability's first slot (slot engine grid)

	reqBody := map[string]interface{}{
		"availability_id":  availability.ID.String(),
		"product_id":       productID.String(),
		"slot_start_time":  slotStartTime.Format(time.RFC3339),
		"guest_first_name": "Alice",
		"guest_last_name":  "Martin",
		"guest_email":      guestEmail,
		"guest_phone":      guestPhone,
	}
	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, testServerURL+"/bookings", bytes.NewReader(bodyBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode, "guest booking creation must succeed")

	var created domain.BookingResponse
	err = json.NewDecoder(resp.Body).Decode(&created)
	require.NoError(t, err)
	require.NotEmpty(t, created.Token, "booking token must be present when WithTokenSecret is configured")
	// A new booking awaits payment until Stripe confirms it (ADR-0018).
	confirmBookingPayment(t, ctx, created.ID)

	return created.ID, created.Token
}

func TestLookupBooking(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("token path - valid token returns booking details", func(t *testing.T) {
		defer tu.ClearAuthData(t, context.Background(), authCtx)

		bookingID, token := setupGuestBookingForLookup(t)

		req, err := http.NewRequest(http.MethodGet,
			testServerURL+"/bookings/lookup?token="+token, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.PublicBookingLookupResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, bookingID, result.ID)
		assert.Equal(t, domain.BookingStatusConfirmed, result.Status)
		// Guest contact fields must never appear in the public response
		raw, _ := json.Marshal(result)
		assert.NotContains(t, string(raw), guestEmail)
		assert.NotContains(t, string(raw), guestPhone)
	})

	t.Run("token path - expired token returns 401", func(t *testing.T) {
		defer tu.ClearAuthData(t, context.Background(), authCtx)

		// Generate a token whose slot end time is in the past (already expired)
		pastSlotEnd := time.Now().Add(-domain.BookingTokenGracePeriod - time.Hour)
		expiredToken := domain.GenerateBookingToken(uuid.New(), pastSlotEnd, testTokenSecret)

		req, err := http.NewRequest(http.MethodGet,
			testServerURL+"/bookings/lookup?token="+expiredToken, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("token path - tampered token returns 401", func(t *testing.T) {
		defer tu.ClearAuthData(t, context.Background(), authCtx)

		req, err := http.NewRequest(http.MethodGet,
			testServerURL+"/bookings/lookup?token=not-a-valid-token", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("manual fallback - valid ref + email returns booking details", func(t *testing.T) {
		defer tu.ClearAuthData(t, context.Background(), authCtx)

		bookingID, _ := setupGuestBookingForLookup(t)

		req, err := http.NewRequest(http.MethodGet,
			fmt.Sprintf("%s/bookings/lookup?ref=%s&email=%s", testServerURL, bookingID, guestEmail), nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.PublicBookingLookupResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, bookingID, result.ID)
	})

	t.Run("manual fallback - valid ref + phone returns booking details", func(t *testing.T) {
		defer tu.ClearAuthData(t, context.Background(), authCtx)

		bookingID, _ := setupGuestBookingForLookup(t)

		req, err := http.NewRequest(http.MethodGet,
			fmt.Sprintf("%s/bookings/lookup?ref=%s&phone=%s", testServerURL, bookingID, guestPhone), nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.PublicBookingLookupResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, bookingID, result.ID)
	})

	t.Run("manual fallback - wrong booking reference returns 401", func(t *testing.T) {
		defer tu.ClearAuthData(t, context.Background(), authCtx)

		wrongRef := uuid.New()
		req, err := http.NewRequest(http.MethodGet,
			fmt.Sprintf("%s/bookings/lookup?ref=%s&email=%s", testServerURL, wrongRef, guestEmail), nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("manual fallback - mismatched contact returns 401", func(t *testing.T) {
		defer tu.ClearAuthData(t, context.Background(), authCtx)

		bookingID, _ := setupGuestBookingForLookup(t)

		req, err := http.NewRequest(http.MethodGet,
			fmt.Sprintf("%s/bookings/lookup?ref=%s&email=wrong@example.com", testServerURL, bookingID), nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("missing params returns 400", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, testServerURL+"/bookings/lookup", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// switchablePracticeAccess is the in-process practice access provider the
// suite's booking service reads, made to fail while fail is set.
type switchablePracticeAccess struct {
	provider bookingPorts.PracticeAccessProvider
	fail     atomic.Bool
}

func (p *switchablePracticeAccess) GetPracticeAccess(ctx context.Context, roomID uuid.UUID) (domain.PracticeAccess, error) {
	if p.fail.Load() {
		return domain.PracticeAccess{}, errors.New("practice access unavailable")
	}
	return p.provider.GetPracticeAccess(ctx, roomID)
}

const (
	testDoorCode           = "4821B"
	testAccessInstructions = "Interphone « Cabinet Leviosa »\n2e étage, porte de gauche"
	testCompanyAddress     = "8 avenue de l'Opéra, 75001 Paris"
	// testBuildingAddress is the address of tb.NewTestBuilding, as the
	// fallback formats it.
	testBuildingAddress = "123 Rue de Rivoli, 75001 Paris"
)

// setPracticeAccess stores the door code and instructions through the real
// settings service, as the admin settings endpoint does.
func setPracticeAccess(t *testing.T, code, instructions string) {
	t.Helper()
	_, err := settingsSvc.SetPracticeAccess(context.Background(), &settingsDomain.SetPracticeAccessRequest{DoorCode: code, AccessInstructions: instructions})
	require.NoError(t, err)
}

func setCompanyAddress(t *testing.T, address string) {
	t.Helper()
	_, err := settingsSvc.SetCompanyLegalAddress(context.Background(), &settingsDomain.SetCompanyLegalAddressRequest{Address: address})
	require.NoError(t, err)
}

func clearCompanyAddress(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "DELETE FROM settings.plain WHERE key = $1", settingsContracts.CompanyLegalAddress)
	require.NoError(t, err)
}

// lookupBothPaths looks the guest booking up by token and by reference +
// email, and returns both responses with their raw bodies.
func lookupBothPaths(t *testing.T, bookingID uuid.UUID, token string) map[string]lookupResult {
	t.Helper()
	return map[string]lookupResult{
		"token":       lookup(t, testServerURL+"/bookings/lookup?token="+token),
		"ref+contact": lookup(t, fmt.Sprintf("%s/bookings/lookup?ref=%s&email=%s", testServerURL, bookingID, guestEmail)),
	}
}

type lookupResult struct {
	booking domain.PublicBookingLookupResponse
	raw     string
}

func lookup(t *testing.T, url string) lookupResult {
	t.Helper()
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	var result lookupResult
	require.NoError(t, json.Unmarshal(body, &result.booking))
	result.raw = string(body)
	return result
}

func assertNoAccess(t *testing.T, results map[string]lookupResult) {
	t.Helper()
	for path, r := range results {
		assert.Nil(t, r.booking.Access, path)
		assert.NotContains(t, r.raw, `"access"`, path)
	}
}

// TestLookupBooking_PracticeAccess: the booking page shows the practice
// access only while the booking is confirmed and its slot has not ended
// (docs/issues/024-practice-access-on-booking-page.md).
func TestLookupBooking_PracticeAccess(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		setPracticeAccess(t, "", "")
		clearCompanyAddress(t)
	})

	t.Run("a confirmed upcoming booking carries the access through both paths", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		setPracticeAccess(t, testDoorCode, testAccessInstructions)
		setCompanyAddress(t, testCompanyAddress)

		bookingID, token := setupGuestBookingForLookup(t)

		for path, r := range lookupBothPaths(t, bookingID, token) {
			require.NotNil(t, r.booking.Access, path)
			assert.Equal(t, domain.PracticeAccessView{
				Address:      testCompanyAddress,
				DoorCode:     testDoorCode,
				Instructions: testAccessInstructions,
			}, *r.booking.Access, path)
		}
	})

	t.Run("a door code change shows on the next lookup", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		setPracticeAccess(t, testDoorCode, testAccessInstructions)
		setCompanyAddress(t, testCompanyAddress)

		bookingID, token := setupGuestBookingForLookup(t)
		before := lookup(t, testServerURL+"/bookings/lookup?token="+token)
		require.NotNil(t, before.booking.Access)
		assert.Equal(t, testDoorCode, before.booking.Access.DoorCode)

		setPracticeAccess(t, "9073", testAccessInstructions)

		for path, r := range lookupBothPaths(t, bookingID, token) {
			require.NotNil(t, r.booking.Access, path)
			assert.Equal(t, "9073", r.booking.Access.DoorCode, path)
		}
	})

	t.Run("the address falls back to the building's when the company address is unset", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		setPracticeAccess(t, testDoorCode, "")
		clearCompanyAddress(t)

		bookingID, token := setupGuestBookingForLookup(t)

		for path, r := range lookupBothPaths(t, bookingID, token) {
			require.NotNil(t, r.booking.Access, path)
			assert.Equal(t, testBuildingAddress, r.booking.Access.Address, path)
			assert.Equal(t, testDoorCode, r.booking.Access.DoorCode, path)
			assert.Empty(t, r.booking.Access.Instructions, path)
		}
	})

	t.Run("instructions alone are enough to carry the access", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		setPracticeAccess(t, "", testAccessInstructions)
		setCompanyAddress(t, testCompanyAddress)

		bookingID, token := setupGuestBookingForLookup(t)

		for path, r := range lookupBothPaths(t, bookingID, token) {
			require.NotNil(t, r.booking.Access, path)
			assert.Empty(t, r.booking.Access.DoorCode, path)
			assert.Equal(t, testAccessInstructions, r.booking.Access.Instructions, path)
		}
	})

	t.Run("no access when neither a door code nor instructions are set", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		setPracticeAccess(t, "", "")
		setCompanyAddress(t, testCompanyAddress)

		bookingID, token := setupGuestBookingForLookup(t)

		assertNoAccess(t, lookupBothPaths(t, bookingID, token))
	})

	for _, status := range []domain.BookingStatus{
		domain.BookingStatusAwaitingPayment,
		domain.BookingStatusExpired,
		domain.BookingStatusCancelled,
		domain.BookingStatusCompleted,
		domain.BookingStatusNoShow,
	} {
		t.Run("no access for a "+string(status)+" booking", func(t *testing.T) {
			defer tu.ClearAuthData(t, ctx, authCtx)
			setPracticeAccess(t, testDoorCode, testAccessInstructions)

			bookingID, token := setupGuestBookingForLookup(t)
			require.NoError(t, tbooking.UpdateBookingStatus(t, ctx, testPool, bookingID, status))

			results := lookupBothPaths(t, bookingID, token)
			for path, r := range results {
				assert.Equal(t, status, r.booking.Status, path)
				assert.NotContains(t, r.raw, testDoorCode, path)
			}
			assertNoAccess(t, results)
		})
	}

	t.Run("no access once the slot has ended", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		setPracticeAccess(t, testDoorCode, testAccessInstructions)

		bookingID, token := setupGuestBookingForLookup(t)
		// The slot starts in about a day: four days on, it has ended.
		suiteClockOffset.Store(int64(4 * 24 * time.Hour))
		defer suiteClockOffset.Store(0)

		results := lookupBothPaths(t, bookingID, token)
		for path, r := range results {
			assert.Equal(t, domain.BookingStatusConfirmed, r.booking.Status, path)
		}
		assertNoAccess(t, results)
	})

	t.Run("a failing access read still returns the booking, without access, and logs no value", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		setPracticeAccess(t, testDoorCode, testAccessInstructions)
		setCompanyAddress(t, testCompanyAddress)

		bookingID, token := setupGuestBookingForLookup(t)

		var logs bytes.Buffer
		previous := slog.Default()
		slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
		defer slog.SetDefault(previous)
		practiceAccess.fail.Store(true)
		defer practiceAccess.fail.Store(false)

		results := lookupBothPaths(t, bookingID, token)
		for path, r := range results {
			assert.Equal(t, bookingID, r.booking.ID, path)
		}
		assertNoAccess(t, results)

		assert.Contains(t, logs.String(), "failed to read the practice access")
		assert.NotContains(t, logs.String(), testDoorCode)
		assert.NotContains(t, logs.String(), "2e étage")
	})
}
