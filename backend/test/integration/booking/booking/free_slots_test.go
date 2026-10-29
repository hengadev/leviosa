package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	settingsContracts "github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	settingsDomain "github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
	th "github.com/Leviosa-care/leviosa/backend/test/helpers"
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

// The public slot endpoint and booking creation run the same slot engine
// (docs/issues/010-slot-engine.md): the page offers exactly what the server
// accepts.
//
// make test-func TEST_NAME=TestFreeSlots TEST_PATH=test/integration/booking/booking/free_slots_test.go

type slotFixture struct {
	partnerID      uuid.UUID // public partner profile ID (what the settings hold)
	partnerUserID  uuid.UUID
	productID      uuid.UUID
	availabilityID uuid.UUID
	day            time.Time // the availability's date, midnight Europe/Paris
	clientID       uuid.UUID
}

// at returns the given wall-clock time on the fixture's day in the practice time zone.
func (f slotFixture) at(hour, minute int) time.Time {
	return time.Date(f.day.Year(), f.day.Month(), f.day.Day(), hour, minute, 0, 0, domain.PracticeLocation)
}

// setupSlotFixture creates a partner with a one-off availability from 09:00
// to 13:00 (Europe/Paris) three days from now, and the 60-minute product with
// a 15-minute buffer: the engine's grid is 09:00, 10:15, 11:30 (12:45 would
// end after 13:00).
func setupSlotFixture(t *testing.T, ctx context.Context) slotFixture {
	t.Helper()

	tbooking.ClearBookingsTable(t, ctx, testPool)
	tavail.ClearAvailabilityTable(t, ctx, testPool)
	tr.ClearRoomsTable(t, ctx, testPool)
	tb.ClearBuildingsTable(t, ctx, testPool)
	tcatalog.ClearProductsTable(t, ctx, testPool)

	categoryID := tcatalog.CreateTestCategory(t, ctx, testPool)
	products := tcatalog.CreateDefaultTestProducts(t, ctx, testPool, categoryID)
	require.Equal(t, 60, products[0].Duration)
	require.Equal(t, 15, products[0].BufferTime)

	building := tb.NewTestBuilding(t)
	buildingEncx, err := domain.ProcessBuildingEncx(ctx, crypto, building)
	require.NoError(t, err)
	require.NoError(t, tb.InsertBuildingEncx(t, ctx, testPool, buildingEncx))

	room := tr.NewTestRoomWithBuilding(t, building.ID)
	roomEncx, err := domain.ProcessRoomEncx(ctx, crypto, room)
	require.NoError(t, err)
	require.NoError(t, tr.InsertRoomEncx(t, ctx, testPool, roomEncx))

	_, partnerUserID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "partner-slots@example.com", room.ID, testPool, authCtx.Redis, crypto)
	partnerEncx, err := th.GetPartnerEncxByUserID(t, ctx, partnerUserID, testPool)
	require.NoError(t, err)
	_, clientID := tsetup.SetupStandardUser(t, ctx, "client-slots@example.com", room.ID, testPool, authCtx.Redis, crypto)

	inThreeDays := time.Now().In(domain.PracticeLocation).AddDate(0, 0, 3)
	fixture := slotFixture{
		partnerID:     partnerEncx.ID,
		partnerUserID: partnerUserID,
		productID:     products[0].ID,
		day:           time.Date(inThreeDays.Year(), inThreeDays.Month(), inThreeDays.Day(), 0, 0, 0, 0, domain.PracticeLocation),
		clientID:      clientID,
	}

	availability := tavail.NewTestAvailability(t, partnerUserID, room.ID, fixture.at(9, 0), fixture.at(13, 0), 0)
	availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
	require.NoError(t, err)
	tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)
	fixture.availabilityID = availability.ID

	return fixture
}

func getFreeSlots(t *testing.T, client *http.Client, partnerID, productID uuid.UUID, from, to string) (int, domain.FreeSlotsResponse) {
	t.Helper()
	url := fmt.Sprintf("%s/slots?partner_id=%s&product_id=%s&from=%s&to=%s", testServerURL, partnerID, productID, from, to)
	resp, err := client.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	var body domain.FreeSlotsResponse
	if resp.StatusCode == http.StatusOK {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	}
	return resp.StatusCode, body
}

func slotStarts(response domain.FreeSlotsResponse) []time.Time {
	starts := make([]time.Time, 0, len(response.Slots))
	for _, slot := range response.Slots {
		starts = append(starts, slot.Start)
	}
	return starts
}

func assertSameInstants(t *testing.T, expected, actual []time.Time) {
	t.Helper()
	require.Len(t, actual, len(expected), "slots: %v", actual)
	for i := range expected {
		assert.True(t, expected[i].Equal(actual[i]), "slot %d: expected %s, got %s", i, expected[i], actual[i])
	}
}

func setBookingPolicy(t *testing.T, ctx context.Context, noticeHours, horizonWeeks int) {
	t.Helper()
	require.NoError(t, settingsRepo.SetInt(ctx, &settingsDomain.Setting[int]{Key: settingsContracts.BookingMinimumNoticeHours, Value: noticeHours}))
	require.NoError(t, settingsRepo.SetInt(ctx, &settingsDomain.Setting[int]{Key: settingsContracts.BookingHorizonWeeks, Value: horizonWeeks}))
}

func postBooking(t *testing.T, client *http.Client, availabilityID, productID uuid.UUID, start time.Time) int {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"availability_id":  availabilityID.String(),
		"product_id":       productID.String(),
		"slot_start_time":  start.Format(time.RFC3339),
		"guest_first_name": "Claire",
		"guest_last_name":  "Dupont",
		"guest_email":      "claire@example.com",
		"guest_phone":      "0611112222",
	})
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, testServerURL+"/bookings", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	return resp.StatusCode
}

func TestFreeSlots(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}
	defer setBookingPolicy(t, ctx, suiteMinimumNoticeHours, settingsDomain.DefaultBookingHorizonWeeks)

	t.Run("should list the engine's slots for a one-off availability, without authentication", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)
		day := f.day.Format(time.DateOnly)

		status, body := getFreeSlots(t, client, f.partnerID, f.productID, day, day)
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "Europe/Paris", body.TimeZone)
		assertSameInstants(t, []time.Time{f.at(9, 0), f.at(10, 15), f.at(11, 30)}, slotStarts(body))
		for _, slot := range body.Slots {
			assert.Equal(t, f.availabilityID, slot.AvailabilityID)
			assert.True(t, slot.End.Equal(slot.Start.Add(60*time.Minute)))
		}

		// A booking at 10:15 takes that slot; its buffer ends at 11:30, so
		// 11:30 stays free.
		_, err := service.CreateBooking(ctx, f.availabilityID, &f.clientID, f.productID, f.at(10, 15), "", "", "", "")
		require.NoError(t, err)

		status, body = getFreeSlots(t, client, f.partnerID, f.productID, day, day)
		require.Equal(t, http.StatusOK, status)
		assertSameInstants(t, []time.Time{f.at(9, 0), f.at(11, 30)}, slotStarts(body))
	})

	t.Run("should apply the booking policy from the settings at once", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)
		day := f.day.Format(time.DateOnly)

		// A one-week horizon still reaches the day three days out.
		setBookingPolicy(t, ctx, 0, 1)
		status, body := getFreeSlots(t, client, f.partnerID, f.productID, day, day)
		require.Equal(t, http.StatusOK, status)
		assert.Len(t, body.Slots, 3)
		assert.WithinDuration(t, time.Now().Add(7*24*time.Hour), body.HorizonEnd, time.Minute)

		// A minimum notice of a week pushes every slot of that day out of reach.
		setBookingPolicy(t, ctx, 168, 8)
		status, body = getFreeSlots(t, client, f.partnerID, f.productID, day, day)
		require.Equal(t, http.StatusOK, status)
		assert.Empty(t, body.Slots)
		assert.WithinDuration(t, time.Now().Add(168*time.Hour), body.EarliestStart, time.Minute)

		// And booking creation refuses those starts too.
		assert.Equal(t, http.StatusBadRequest, postBooking(t, client, f.availabilityID, f.productID, f.at(9, 0)))

		setBookingPolicy(t, ctx, suiteMinimumNoticeHours, settingsDomain.DefaultBookingHorizonWeeks)
	})

	t.Run("should reject invalid queries", func(t *testing.T) {
		day := time.Now().In(domain.PracticeLocation).Format(time.DateOnly)

		status, _ := getFreeSlots(t, client, uuid.New(), uuid.New(), "not-a-date", day)
		assert.Equal(t, http.StatusBadRequest, status)

		status, _ = getFreeSlots(t, client, uuid.New(), uuid.New(), day, "2000-01-01")
		assert.Equal(t, http.StatusBadRequest, status, "to before from")

		resp, err := client.Get(testServerURL + "/slots?partner_id=nope&product_id=" + uuid.NewString() + "&from=" + day + "&to=" + day)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should answer 404 for an unknown partner", func(t *testing.T) {
		day := time.Now().In(domain.PracticeLocation).Format(time.DateOnly)
		status, _ := getFreeSlots(t, client, uuid.New(), uuid.New(), day, day)
		assert.Equal(t, http.StatusNotFound, status)
	})
}

func TestCreateBookingValidatesWithTheSlotEngine(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}
	defer setBookingPolicy(t, ctx, suiteMinimumNoticeHours, settingsDomain.DefaultBookingHorizonWeeks)

	t.Run("should accept a start the engine offers", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)
		assert.Equal(t, http.StatusCreated, postBooking(t, client, f.availabilityID, f.productID, f.at(11, 30)))
	})

	t.Run("should reject a start off the engine's grid", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)
		// 10:00 is on the old 10-minute grid but not on the slot grid (every 75 min).
		assert.Equal(t, http.StatusBadRequest, postBooking(t, client, f.availabilityID, f.productID, f.at(10, 0)))
	})

	t.Run("should reject a slot that would end after the availability", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)
		assert.Equal(t, http.StatusBadRequest, postBooking(t, client, f.availabilityID, f.productID, f.at(12, 45)))
	})

	t.Run("should reject a start before the minimum notice", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)
		setBookingPolicy(t, ctx, 96, 8) // four days: the fixture's day is three days out
		defer setBookingPolicy(t, ctx, suiteMinimumNoticeHours, settingsDomain.DefaultBookingHorizonWeeks)
		assert.Equal(t, http.StatusBadRequest, postBooking(t, client, f.availabilityID, f.productID, f.at(9, 0)))
	})

	t.Run("should answer a conflict for a slot already booked", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)
		require.Equal(t, http.StatusCreated, postBooking(t, client, f.availabilityID, f.productID, f.at(10, 15)))
		assert.Equal(t, http.StatusConflict, postBooking(t, client, f.availabilityID, f.productID, f.at(10, 15)))
	})
}
