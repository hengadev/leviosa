package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// The slot engine's contract tests (issue 010): behaviour is checked through
// the public function only — given availabilities, occupancies, a policy and
// a clock, which slots come out. Times are UTC instants: the engine works on
// instants, and the practice time zone only matters to callers turning dates
// into ranges (and to recurrence expansion, issue 011).

// permissiveHorizon keeps the policy out of the way in tests that do not
// exercise notice or horizon edges.
const permissiveHorizon = 365 * 24 * time.Hour

func slotEngineFixture() (now, from, to time.Time) {
	// now: Monday 2026-11-09 08:00 UTC; query range: that whole day.
	now = time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)
	from = time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)
	to = from.Add(24 * time.Hour)
	return now, from, to
}

func window(dayStart, dayEnd time.Time) AvailabilityWindow {
	return AvailabilityWindow{Start: dayStart, End: dayEnd}
}

func starts(slots []Slot) []time.Time {
	out := make([]time.Time, 0, len(slots))
	for _, slot := range slots {
		out = append(out, slot.Start)
	}
	return out
}

func at(hour, minute int) time.Time {
	return time.Date(2026, 11, 9, hour, minute, 0, 0, time.UTC)
}

func TestComputeSlotsGridWithoutBuffer(t *testing.T) {
	now, from, to := slotEngineFixture()

	// 60-minute product in a 09:00–12:00 window: 09:00, 10:00, 11:00.
	slots := ComputeSlots(SlotEngineQuery{
		Availabilities:  []AvailabilityWindow{window(at(9, 0), at(12, 0))},
		ProductDuration: 60 * time.Minute,
		BookingHorizon:  permissiveHorizon,
		From:            from,
		To:              to,
		Now:             now,
	})

	assert.Equal(t, []time.Time{at(9, 0), at(10, 0), at(11, 0)}, starts(slots))
	for _, slot := range slots {
		assert.Equal(t, slot.Start.Add(60*time.Minute), slot.End)
	}
}

func TestComputeSlotsGridWithBuffer(t *testing.T) {
	now, from, to := slotEngineFixture()

	// 60-minute product, 15-minute buffer in 09:00–12:30: the grid steps every
	// 75 minutes: 09:00, 10:15, 11:30 (ending exactly at 12:30). A slot at
	// 12:45 would start after the window, so it is not offered.
	slots := ComputeSlots(SlotEngineQuery{
		Availabilities:  []AvailabilityWindow{window(at(9, 0), at(12, 30))},
		ProductDuration: 60 * time.Minute,
		BufferTime:      15 * time.Minute,
		BookingHorizon:  permissiveHorizon,
		From:            from,
		To:              to,
		Now:             now,
	})

	assert.Equal(t, []time.Time{at(9, 0), at(10, 15), at(11, 30)}, starts(slots))
}

func TestComputeSlotsSlotNotFittingAtEndOfWindow(t *testing.T) {
	now, from, to := slotEngineFixture()

	// 09:00–11:50 with a 60-minute product: 09:00 and 10:00 fit; a slot at
	// 11:00 would end at 12:00, after the window — not offered even though
	// its start is inside the window.
	slots := ComputeSlots(SlotEngineQuery{
		Availabilities:  []AvailabilityWindow{window(at(9, 0), at(11, 50))},
		ProductDuration: 60 * time.Minute,
		BookingHorizon:  permissiveHorizon,
		From:            from,
		To:              to,
		Now:             now,
	})

	assert.Equal(t, []time.Time{at(9, 0), at(10, 0)}, starts(slots))

	// The boundary case: a slot ending exactly at the window's end is offered.
	slots = ComputeSlots(SlotEngineQuery{
		Availabilities:  []AvailabilityWindow{window(at(9, 0), at(11, 0))},
		ProductDuration: 60 * time.Minute,
		BookingHorizon:  permissiveHorizon,
		From:            from,
		To:              to,
		Now:             now,
	})
	assert.Equal(t, []time.Time{at(9, 0), at(10, 0)}, starts(slots))
}

func TestComputeSlotsOverlapWithBooking(t *testing.T) {
	now, from, to := slotEngineFixture()

	base := SlotEngineQuery{
		Availabilities:  []AvailabilityWindow{window(at(9, 0), at(12, 0))},
		ProductDuration: 60 * time.Minute,
		BookingHorizon:  permissiveHorizon,
		From:            from,
		To:              to,
		Now:             now,
	}

	t.Run("booking session blocks overlapping slots", func(t *testing.T) {
		// A confirmed booking 10:00–11:00 removes the 10:00 slot.
		query := base
		query.Occupancies = []SlotOccupancy{{Start: at(10, 0), End: at(11, 0)}}
		assert.Equal(t, []time.Time{at(9, 0), at(11, 0)}, starts(ComputeSlots(query)))
	})

	t.Run("booking buffer extends the blocked interval", func(t *testing.T) {
		// The same booking with a 15-minute buffer (occupancy 10:00–11:15)
		// also removes the 11:00 slot: it would start inside the rest time.
		query := base
		query.Occupancies = []SlotOccupancy{{Start: at(10, 0), End: at(11, 15)}}
		assert.Equal(t, []time.Time{at(9, 0)}, starts(ComputeSlots(query)))
	})

	t.Run("candidate buffer counts against a following booking", func(t *testing.T) {
		// Booking 10:15–11:15 with no buffer, in a 09:00–13:00 window with a
		// 30-minute buffer (grid every 90 minutes: 09:00, 10:30, 12:00). The
		// 09:00 slot's session ends at 10:00, but its own buffer would run
		// until 10:30, inside the booking: it is not offered. 10:30 overlaps
		// the booking itself; 12:00 is free.
		query := base
		query.Availabilities = []AvailabilityWindow{window(at(9, 0), at(13, 0))}
		query.BufferTime = 30 * time.Minute
		query.Occupancies = []SlotOccupancy{{Start: at(10, 15), End: at(11, 15)}}
		assert.Equal(t, []time.Time{at(12, 0)}, starts(ComputeSlots(query)))
	})

	t.Run("touching intervals do not overlap", func(t *testing.T) {
		// A booking 09:00–10:00 blocks the 09:00 slot only: touching the
		// 10:00 slot's start does not overlap it.
		query := base
		query.Occupancies = []SlotOccupancy{{Start: at(9, 0), End: at(10, 0)}}
		assert.Equal(t, []time.Time{at(10, 0), at(11, 0)}, starts(ComputeSlots(query)))
	})
}

func TestComputeSlotsMinimumNoticeAndHorizonEdges(t *testing.T) {
	now, from, to := slotEngineFixture()
	// now 08:00, policy: 12 h notice, 8 weeks horizon.
	base := SlotEngineQuery{
		Now:             now,
		Availabilities:  []AvailabilityWindow{window(from, to.Add(9*7*24*time.Hour))},
		ProductDuration: 60 * time.Minute,
		MinimumNotice:   12 * time.Hour,
		BookingHorizon:  8 * 7 * 24 * time.Hour,
		From:            from,
		To:              to.Add(9 * 7 * 24 * time.Hour),
	}

	t.Run("slot before now + notice is not offered", func(t *testing.T) {
		slots := ComputeSlots(base)
		for _, slot := range slots {
			assert.True(t, !slot.Start.Before(base.Now.Add(base.MinimumNotice)),
				"slot at %s starts before now + minimum notice", slot.Start)
		}
		// The first offered slot is exactly now + 12 h (20:00): the boundary
		// itself is offered.
		assert.Equal(t, time.Date(2026, 11, 9, 20, 0, 0, 0, time.UTC), slots[0].Start)
	})

	t.Run("slot at exactly now + horizon is offered, later is not", func(t *testing.T) {
		horizonEnd := base.Now.Add(base.BookingHorizon) // 2027-01-04 08:00 UTC
		query := base
		query.From = horizonEnd.Add(-2 * time.Hour)
		query.To = horizonEnd.Add(2 * time.Hour)

		slots := ComputeSlots(query)
		// 06:00, 07:00 and 08:00 (= now + horizon) are offered; 09:00, inside
		// the requested range, is past the horizon.
		assert.Equal(t, []time.Time{horizonEnd.Add(-2 * time.Hour), horizonEnd.Add(-time.Hour), horizonEnd}, starts(slots))
	})

	t.Run("no slots when the notice window eats the whole range", func(t *testing.T) {
		now := time.Date(2026, 11, 9, 23, 30, 0, 0, time.UTC)
		query := base
		query.Now = now
		query.From = from
		query.To = to // the range ends 00:00, before now + 12 h
		assert.Empty(t, ComputeSlots(query))
	})
}

func TestComputeSlotsOrderingAndWindow(t *testing.T) {
	now, from, to := slotEngineFixture()

	t.Run("slots across windows are ordered by start", func(t *testing.T) {
		// Windows given out of order; the morning window fits one slot (a
		// second at 10:00 would end after 10:30), the afternoon window two.
		slots := ComputeSlots(SlotEngineQuery{
			Availabilities: []AvailabilityWindow{
				window(at(15, 0), at(17, 0)),
				window(at(9, 0), at(10, 30)),
			},
			ProductDuration: 60 * time.Minute,
			BookingHorizon:  permissiveHorizon,
			From:            from,
			To:              to,
			Now:             now,
		})
		assert.Equal(t, []time.Time{at(9, 0), at(15, 0), at(16, 0)}, starts(slots))
	})

	t.Run("only slots starting within From (inclusive) and To (exclusive)", func(t *testing.T) {
		slots := ComputeSlots(SlotEngineQuery{
			Availabilities:  []AvailabilityWindow{window(at(9, 0), at(12, 0))},
			ProductDuration: 60 * time.Minute,
			BookingHorizon:  permissiveHorizon,
			From:            at(10, 0),
			To:              at(11, 0),
			Now:             now,
		})
		assert.Equal(t, []time.Time{at(10, 0)}, starts(slots))
	})

	t.Run("a window overlapping the range only contributes inside it", func(t *testing.T) {
		// Availability 08:00–10:00 that began before the requested range: the
		// grid stays anchored to 08:00, so 09:00 and 09:30 are offered.
		slots := ComputeSlots(SlotEngineQuery{
			Availabilities:  []AvailabilityWindow{window(at(8, 0), at(10, 0))},
			ProductDuration: 30 * time.Minute,
			BookingHorizon:  permissiveHorizon,
			From:            at(9, 0),
			To:              to,
			Now:             now,
			MinimumNotice:   0,
		})
		assert.Equal(t, []time.Time{at(9, 0), at(9, 30)}, starts(slots))
	})
}

func TestComputeSlotsRejectsImpossibleQueries(t *testing.T) {
	now, from, to := slotEngineFixture()

	assert.Nil(t, ComputeSlots(SlotEngineQuery{
		Availabilities:  []AvailabilityWindow{window(at(9, 0), at(12, 0))},
		ProductDuration: 0,
		BookingHorizon:  permissiveHorizon,
		From:            from,
		To:              to,
		Now:             now,
	}), "zero duration can never produce a slot")

	assert.Nil(t, ComputeSlots(SlotEngineQuery{
		Availabilities:  []AvailabilityWindow{window(at(9, 0), at(12, 0))},
		ProductDuration: 60 * time.Minute,
		BufferTime:      -90 * time.Minute,
		BookingHorizon:  permissiveHorizon,
		From:            from,
		To:              to,
		Now:             now,
	}), "negative step would loop forever; it yields no slots instead")
}

func TestPracticeLocation(t *testing.T) {
	// Slots are computed in Europe/Paris (PRD "Slot engine"): the location
	// must resolve from the embedded tzdata, and Paris is UTC+1 in winter.
	assert.Equal(t, "Europe/Paris", PracticeLocation.String())
	winter := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	_, offset := winter.In(PracticeLocation).Zone()
	assert.Equal(t, 3600, offset)
}

func TestComputeSlotsCarriesTheAvailability(t *testing.T) {
	now, from, to := slotEngineFixture()
	morning := uuid.New()
	afternoon := uuid.New()

	slots := ComputeSlots(SlotEngineQuery{
		Availabilities: []AvailabilityWindow{
			{AvailabilityID: afternoon, Start: at(14, 0), End: at(15, 0)},
			{AvailabilityID: morning, Start: at(9, 0), End: at(10, 0)},
		},
		ProductDuration: 60 * time.Minute,
		BookingHorizon:  permissiveHorizon,
		From:            from,
		To:              to,
		Now:             now,
	})

	assert.Equal(t, []Slot{
		{AvailabilityID: morning, Start: at(9, 0), End: at(10, 0)},
		{AvailabilityID: afternoon, Start: at(14, 0), End: at(15, 0)},
	}, slots)
}

func TestOffersSlot(t *testing.T) {
	now, _, _ := slotEngineFixture()
	query := SlotEngineQuery{
		Availabilities:  []AvailabilityWindow{window(at(9, 0), at(13, 0))},
		ProductDuration: 60 * time.Minute,
		BufferTime:      15 * time.Minute,
		Occupancies:     []SlotOccupancy{{Start: at(10, 15), End: at(11, 30)}},
		BookingHorizon:  permissiveHorizon,
		Now:             now,
	}

	assert.True(t, OffersSlot(query, at(9, 0)), "first grid point is offered")
	assert.False(t, OffersSlot(query, at(10, 15)), "grid point taken by a booking")
	assert.True(t, OffersSlot(query, at(11, 30)), "grid point after the booking")
	assert.False(t, OffersSlot(query, at(9, 30)), "off the grid")
	assert.False(t, OffersSlot(query, at(9, 0).Add(time.Second)), "off the grid by a second")

	query.MinimumNotice = 2 * time.Hour // now 08:00: nothing before 10:00
	assert.False(t, OffersSlot(query, at(9, 0)), "inside the minimum notice")
}
