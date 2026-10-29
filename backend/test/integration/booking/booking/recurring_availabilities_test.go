package booking

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	settingsDomain "github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
	tavail "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/availability"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Recurring availabilities (docs/issues/011-recurring-availabilities.md):
// a series created in the partner area offers slots on every occurrence up
// to the booking horizon; editing or deleting it never changes the bookings
// already made; a booking blocks its time in every availability of the
// partner.
//
// make test-func TEST_NAME=TestRecurringAvailabilities TEST_PATH=test/integration/booking/booking/recurring_availabilities_test.go

// onDay returns the wall-clock time days after the fixture's day, in the
// practice time zone.
func (f slotFixture) onDay(days, hour, minute int) time.Time {
	d := f.day.AddDate(0, 0, days)
	return time.Date(d.Year(), d.Month(), d.Day(), hour, minute, 0, 0, domain.PracticeLocation)
}

// createWeeklyAvailability creates, as the partner area does, a weekly
// 09:00–13:00 series on the fixture day's weekday, starting on that day.
func createWeeklyAvailability(t *testing.T, ctx context.Context, f slotFixture) *domain.Availability {
	t.Helper()
	availability, err := availabilitySvc.CreateRecurringAvailability(ctx, &domain.CreateRecurringAvailabilityRequest{
		UserID:      f.partnerUserID,
		RoomID:      f.roomID,
		StartTime:   f.at(9, 0),
		EndTime:     f.at(13, 0),
		MaxCapacity: 1,
		Pattern: domain.RecurrencePattern{
			Type:       domain.RecurrenceWeekly,
			Interval:   1,
			DaysOfWeek: []int{int(f.day.Weekday())},
		},
	})
	require.NoError(t, err)
	return availability
}

func slotsOn(t *testing.T, client *http.Client, f slotFixture, day time.Time) []time.Time {
	t.Helper()
	date := day.In(domain.PracticeLocation).Format(time.DateOnly)
	status, body := getFreeSlots(t, client, f.partnerID, f.productID, date, date)
	require.Equal(t, http.StatusOK, status)
	return slotStarts(body)
}

// bookingSnapshot is what must not change on a booking when its
// availability is edited or deleted.
type bookingSnapshot struct {
	AvailabilityID uuid.UUID
	Status         domain.BookingStatus
	Start, End     time.Time
}

func snapshotBooking(t *testing.T, ctx context.Context, id uuid.UUID) bookingSnapshot {
	t.Helper()
	bookingEncx, err := bookingRepo.GetByID(ctx, id)
	require.NoError(t, err)
	booking, err := domain.DecryptBookingEncx(ctx, crypto, bookingEncx)
	require.NoError(t, err)
	return bookingSnapshot{
		AvailabilityID: booking.AvailabilityID,
		Status:         booking.Status,
		Start:          booking.SlotStartTime.UTC(),
		End:            booking.SlotEndTime.UTC(),
	}
}

func TestRecurringAvailabilities(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}
	defer setBookingPolicy(t, ctx, suiteMinimumNoticeHours, settingsDomain.DefaultBookingHorizonWeeks)

	t.Run("should offer a weekly availability's slots every week up to the booking horizon", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotPartner(t, ctx)
		setBookingPolicy(t, ctx, 0, 4)
		availability := createWeeklyAvailability(t, ctx, f)

		// The calendar asks for five weeks; the horizon (four weeks from now)
		// keeps the occurrences three days, one, two and three weeks out.
		from := time.Now().In(domain.PracticeLocation).Format(time.DateOnly)
		to := f.onDay(35, 0, 0).Format(time.DateOnly)
		status, body := getFreeSlots(t, client, f.partnerID, f.productID, from, to)
		require.Equal(t, http.StatusOK, status)

		var expected []time.Time
		for week := range 4 {
			days := 7 * week
			expected = append(expected, f.onDay(days, 9, 0), f.onDay(days, 10, 15), f.onDay(days, 11, 30))
		}
		assertSameInstants(t, expected, slotStarts(body))
		for _, slot := range body.Slots {
			assert.Equal(t, availability.ID, slot.AvailabilityID)
			assert.Equal(t, f.day.Weekday(), slot.Start.Weekday(), "times are rendered in Europe/Paris")
		}
	})

	t.Run("should book any occurrence of a recurring availability and block only that one", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotPartner(t, ctx)
		availability := createWeeklyAvailability(t, ctx, f)

		// The series' first occurrence is three days out; a booking one week
		// later is validated against that week's occurrence.
		require.Equal(t, http.StatusCreated, postBooking(t, client, availability.ID, f.productID, f.onDay(7, 10, 15)))
		assert.Equal(t, http.StatusConflict, postBooking(t, client, availability.ID, f.productID, f.onDay(7, 10, 15)))
		assert.Equal(t, http.StatusBadRequest, postBooking(t, client, availability.ID, f.productID, f.onDay(8, 10, 15)),
			"no occurrence on the next day")

		assertSameInstants(t, []time.Time{f.onDay(7, 9, 0), f.onDay(7, 11, 30)}, slotsOn(t, client, f, f.onDay(7, 0, 0)))
		assertSameInstants(t, []time.Time{f.onDay(14, 9, 0), f.onDay(14, 10, 15), f.onDay(14, 11, 30)},
			slotsOn(t, client, f, f.onDay(14, 0, 0))) // the same time a week later stays free
	})

	t.Run("should leave existing bookings unchanged when a recurring availability is edited", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotPartner(t, ctx)
		availability := createWeeklyAvailability(t, ctx, f)

		booked, err := service.CreateBooking(ctx, availability.ID, &f.clientID, f.productID, f.onDay(7, 10, 15), "", "", "", "", "")
		require.NoError(t, err)
		before := snapshotBooking(t, ctx, booked.ID)

		// The partner moves the series to the afternoon, from the fixture's day on.
		newStart, newEnd := f.at(14, 0), f.at(18, 0)
		_, err = availabilitySvc.UpdateAvailability(ctx, &domain.UpdateAvailabilityRequest{
			ID: availability.ID, StartTime: &newStart, EndTime: &newEnd,
		})
		require.NoError(t, err)

		assert.Equal(t, before, snapshotBooking(t, ctx, booked.ID), "the booking keeps its time, status and availability")
		assertSameInstants(t, []time.Time{f.onDay(7, 14, 0), f.onDay(7, 15, 15), f.onDay(7, 16, 30)},
			slotsOn(t, client, f, f.onDay(7, 0, 0))) // the edit applies to the series

		// Moved back over the booking's time, the series does not offer it.
		newStart, newEnd = f.at(10, 15), f.at(13, 0)
		_, err = availabilitySvc.UpdateAvailability(ctx, &domain.UpdateAvailabilityRequest{
			ID: availability.ID, StartTime: &newStart, EndTime: &newEnd,
		})
		require.NoError(t, err)
		assertSameInstants(t, []time.Time{f.onDay(7, 11, 30)}, slotsOn(t, client, f, f.onDay(7, 0, 0)))
		assert.Equal(t, before, snapshotBooking(t, ctx, booked.ID))
	})

	t.Run("should leave existing bookings unchanged when a recurring availability is deleted", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotPartner(t, ctx)
		availability := createWeeklyAvailability(t, ctx, f)

		booked, err := service.CreateBooking(ctx, availability.ID, &f.clientID, f.productID, f.onDay(7, 10, 15), "", "", "", "", "")
		require.NoError(t, err)
		before := snapshotBooking(t, ctx, booked.ID)

		require.NoError(t, availabilitySvc.CancelAvailability(ctx, availability.ID))

		assert.Equal(t, before, snapshotBooking(t, ctx, booked.ID), "the booking keeps its time, status and availability")
		assert.Empty(t, slotsOn(t, client, f, f.onDay(7, 0, 0)), "a deleted series offers no slot")
		assert.Empty(t, slotsOn(t, client, f, f.onDay(14, 0, 0)))
		assert.Equal(t, http.StatusConflict, postBooking(t, client, availability.ID, f.productID, f.onDay(14, 9, 0)))

		stored := tavail.GetAvailabilityEncxFromDB(t, ctx, availability.ID, testPool)
		assert.Equal(t, domain.AvailabilityStatusCancelled, stored.Status, "deleting deactivates the availability, the row stays")
	})

	t.Run("should block a booked time in every overlapping availability of the partner", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotPartner(t, ctx)
		series := createWeeklyAvailability(t, ctx, f)

		// A one-off availability on top of the series' second occurrence
		// (inserted directly: the partner area's conflict check only looks
		// at a series' first occurrence).
		oneOff := tavail.NewTestAvailability(t, f.partnerUserID, f.roomID, f.onDay(7, 10, 15), f.onDay(7, 13, 0), 0)
		oneOffEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, oneOff)
		require.NoError(t, err)
		tavail.InsertAvailabilityEncx(t, ctx, oneOffEncx, testPool)

		// Both availabilities offer 10:15 and 11:30, each start once.
		assertSameInstants(t, []time.Time{f.onDay(7, 9, 0), f.onDay(7, 10, 15), f.onDay(7, 11, 30)},
			slotsOn(t, client, f, f.onDay(7, 0, 0)))

		require.Equal(t, http.StatusCreated, postBooking(t, client, oneOff.ID, f.productID, f.onDay(7, 10, 15)))

		// The booking in the one-off blocks 10:15 in the series too.
		assertSameInstants(t, []time.Time{f.onDay(7, 9, 0), f.onDay(7, 11, 30)}, slotsOn(t, client, f, f.onDay(7, 0, 0)))
		assert.Equal(t, http.StatusConflict, postBooking(t, client, series.ID, f.productID, f.onDay(7, 10, 15)),
			fmt.Sprintf("the series cannot sell the time booked in availability %s", oneOff.ID))
	})
}
