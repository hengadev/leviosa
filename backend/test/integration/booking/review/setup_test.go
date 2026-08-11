package review

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	userDomain "github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	th "github.com/Leviosa-care/leviosa/backend/test/helpers"
	tsetup "github.com/Leviosa-care/leviosa/backend/test/helpers/booking"
	tavail "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/availability"
	tbooking "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/booking"
	tb "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/building"
	tr "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/room"
	tcatalog "github.com/Leviosa-care/leviosa/backend/test/helpers/catalog"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// bookingFixture holds the identities and access tokens produced by newBooking, for
// tests that need to act as the client, partner, or an admin against the resulting
// booking/review.
type bookingFixture struct {
	BookingID    uuid.UUID
	ProductID    uuid.UUID
	PartnerID    uuid.UUID
	ClientID     uuid.UUID
	ClientToken  string
	PartnerToken string
}

// clearReviewTestTables resets every table touched by review test setup, so each
// subtest starts from a clean slate. This includes auth.users/sessions: some setup
// helpers (e.g. SetupAdminWithAllocation) use a fixed email, which would otherwise
// collide across subtests.
func clearReviewTestTables(t *testing.T, ctx context.Context) {
	t.Helper()
	tu.ClearAuthData(t, ctx, authCtx)
	_, err := testPool.Exec(ctx, "TRUNCATE TABLE booking.reviews RESTART IDENTITY CASCADE")
	require.NoError(t, err)
	tbooking.ClearBookingsTable(t, ctx, testPool)
	tavail.ClearAvailabilityTable(t, ctx, testPool)
	tr.ClearRoomsTable(t, ctx, testPool)
	tb.ClearBuildingsTable(t, ctx, testPool)
	tcatalog.ClearProductsTable(t, ctx, testPool)
}

// partnerFixture holds a partner/room/product already set up, so multiple distinct
// bookings can be created against the same partner (e.g. to test rating aggregation,
// which needs several real bookings — booking_id has a foreign key to
// booking.bookings, so reviews can't be backdated against synthetic IDs).
type partnerFixture struct {
	PartnerID    uuid.UUID
	PartnerToken string
	RoomID       uuid.UUID
	ProductID    uuid.UUID
}

// newPartner sets up a building/room/product and an authenticated partner allocated
// to that room.
func newPartner(t *testing.T, ctx context.Context) *partnerFixture {
	t.Helper()

	building := tb.NewTestBuilding(t)
	buildingEncx, err := domain.ProcessBuildingEncx(ctx, crypto, building)
	require.NoError(t, err)
	require.NoError(t, tb.InsertBuildingEncx(t, ctx, testPool, buildingEncx))

	room := tr.NewTestRoomWithBuilding(t, building.ID)
	roomEncx, err := domain.ProcessRoomEncx(ctx, crypto, room)
	require.NoError(t, err)
	require.NoError(t, tr.InsertRoomEncx(t, ctx, testPool, roomEncx))

	categoryID := tcatalog.CreateTestCategory(t, ctx, testPool)
	products := tcatalog.CreateDefaultTestProducts(t, ctx, testPool, categoryID)
	productID := products[0].ID

	partnerToken, partnerID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, uniqueEmail(t, "partner"), room.ID, testPool, authCtx.Redis, crypto)

	return &partnerFixture{
		PartnerID:    partnerID,
		PartnerToken: partnerToken,
		RoomID:       room.ID,
		ProductID:    productID,
	}
}

// newBookingForPartner creates a fresh client and a new availability/booking against
// an already set-up partner. The booking is left in `confirmed` status — callers that
// need a `completed` booking should call completeBooking next.
func newBookingForPartner(t *testing.T, ctx context.Context, client *http.Client, partner *partnerFixture) *bookingFixture {
	t.Helper()

	clientToken, clientID := tsetup.SetupStandardUser(t, ctx, uniqueEmail(t, "client"), partner.RoomID, testPool, authCtx.Redis, crypto)

	// Offset each booking's slot by a few hours so multiple bookings against the same
	// partner/room in one test don't collide on room-time-overlap validation.
	offset := time.Duration(nextSlotOffset()) * 3 * time.Hour
	startTime := time.Now().Add(24*time.Hour + offset).Truncate(time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	availability := tavail.NewTestAvailability(t, partner.PartnerID, partner.RoomID, startTime, endTime, 5000)
	availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
	require.NoError(t, err)
	tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)

	slotStartTime := startTime.Add(30 * time.Minute).Truncate(10 * time.Minute)
	requestBody := map[string]interface{}{
		"availability_id": availability.ID.String(),
		"client_id":       clientID,
		"product_id":      partner.ProductID.String(),
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

	var created domain.BookingResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	return &bookingFixture{
		BookingID:    created.ID,
		ProductID:    partner.ProductID,
		PartnerID:    partner.PartnerID,
		ClientID:     clientID,
		ClientToken:  clientToken,
		PartnerToken: partner.PartnerToken,
	}
}

// newBooking sets up a fresh partner and a single booking against it. The booking is
// left in `confirmed` status — callers that need a `completed` booking should call
// completeBooking next, or use newCompletedBooking.
func newBooking(t *testing.T, ctx context.Context, client *http.Client) *bookingFixture {
	t.Helper()
	return newBookingForPartner(t, ctx, client, newPartner(t, ctx))
}

// newCompletedBooking is newBooking followed by a partner-initiated completion, for
// tests that need a booking already in `completed` status.
func newCompletedBooking(t *testing.T, ctx context.Context, client *http.Client) *bookingFixture {
	t.Helper()

	fixture := newBooking(t, ctx, client)
	completeBooking(t, client, fixture.BookingID, fixture.PartnerToken)
	return fixture
}

// completeBooking marks a booking as completed via HTTP, acting as the partner.
func completeBooking(t *testing.T, client *http.Client, bookingID uuid.UUID, partnerToken string) {
	t.Helper()

	req, err := http.NewRequest("POST", testServerURL+"/bookings/"+bookingID.String()+"/complete", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: partnerToken})

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// newAdminToken creates an admin session for moderation tests. Review moderation
// (hide/unhide) only checks role, not room allocation, so — unlike
// tsetup.SetupAdminWithAllocation — this does not require a real room row to satisfy
// a foreign key.
func newAdminToken(t *testing.T, ctx context.Context) string {
	t.Helper()

	userID := uuid.New()
	user := th.NewTestUser(t, uniqueEmail(t, "admin"), "Admin", "User")
	user.ID = userID
	user.Role = identity.Administrator.String()

	userEncx, err := userDomain.ProcessUserEncx(ctx, crypto, user)
	require.NoError(t, err)
	require.NoError(t, th.InsertUserEncx(t, ctx, userEncx, testPool))

	return tsetup.CreateSessionForUser(t, ctx, userID, identity.Administrator, authCtx.Redis, crypto)
}

// uniqueEmail avoids collisions across subtests that don't clear the users table
// between runs.
func uniqueEmail(t *testing.T, prefix string) string {
	t.Helper()
	return prefix + "-" + uuid.New().String()[:8] + "@test.com"
}

var slotOffsetCounter int64

// nextSlotOffset returns a monotonically increasing counter, used to space out
// booking slots created within a single test so they don't overlap.
func nextSlotOffset() int64 {
	return atomic.AddInt64(&slotOffsetCounter, 1)
}
