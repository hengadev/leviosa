// Package domain: slot engine.
//
// The slot engine derives the bookable start times (slots) a client may pick,
// from the partner's availabilities, the product's duration and buffer time,
// the bookings already occupying time, and the booking policy. It is a pure
// function: no repositories, no clock, no infrastructure — callers pass "now"
// and every input. The practice page's calendar and booking creation both run
// it, so the two can never disagree (docs/issues/010-slot-engine.md, PRD
// "Slot engine").
//
// Its input is occurrence-shaped on purpose: `AvailabilityWindow` is one
// occurrence of an availability. Recurring availabilities (issue 011) are
// expanded into occurrences by the caller (Availability.Occurrences,
// recurrence.go) before they reach the engine, and
// availability exceptions (issue 012) remove occurrences the same way — the
// engine itself never needs to know about recurrence. Occupancies are plain
// intervals so held bookings (issue 014) join without a signature change.
package domain

import (
	"sort"
	"time"

	"github.com/google/uuid"

	// Embed the time zone database: the practice location below must resolve
	// even on scratch containers without system tzdata.
	_ "time/tzdata"
)

// PracticeLocation is the time zone the practice operates in. Occurrences are
// expressed in it (a weekly 09:00 stays 09:00 across daylight-saving changes),
// and date ranges from the public API are parsed in it.
var PracticeLocation = loadPracticeLocation()

func loadPracticeLocation() *time.Location {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		// Unreachable: the tzdata database is embedded above.
		panic("slot engine: cannot load Europe/Paris: " + err.Error())
	}
	return loc
}

// Slot is one bookable start time with its end (start + product duration),
// and the availability it belongs to (a booking is created against it).
type Slot struct {
	AvailabilityID uuid.UUID
	Start          time.Time
	End            time.Time
}

// AvailabilityWindow is one occurrence of a partner availability: the partner
// is open for bookings between Start and End. One-off availabilities are their
// own single occurrence; recurring ones are expanded by the caller.
type AvailabilityWindow struct {
	AvailabilityID uuid.UUID
	Start          time.Time
	End            time.Time
}

// SlotOccupancy is a time interval already taken by a booking, the booking's
// buffer included: nothing that needs the room may be offered inside it. The
// caller derives it from an occupying booking (`confirmed`, and an
// `awaiting_payment` booking whose hold has not expired) by extending the
// booking's end with the booked product's buffer time.
type SlotOccupancy struct {
	Start time.Time
	End   time.Time
}

// SlotEngineQuery is the slot engine's complete input.
type SlotEngineQuery struct {
	// Availabilities are the occurrence windows to step through, in any order.
	Availabilities []AvailabilityWindow

	// ProductDuration is the session length; BufferTime is the mandatory rest
	// after a session before the room may be used again. Slots start at each
	// occurrence's start and follow each other every ProductDuration + BufferTime.
	ProductDuration time.Duration
	BufferTime      time.Duration

	// Occupancies are the intervals blocked by existing bookings.
	Occupancies []SlotOccupancy

	// Booking policy: a slot starting before Now+MinimumNotice or after
	// Now+BookingHorizon is not offered.
	MinimumNotice  time.Duration
	BookingHorizon time.Duration

	// Now is the moment the query is evaluated for.
	Now time.Time

	// From (inclusive) and To (exclusive) bound the returned slots' starts.
	// Callers pass the requested date range clamped to the policy window.
	From time.Time
	To   time.Time
}

// ComputeSlots returns the free slots for the query, ordered by start time.
//
// Rules (PRD "Slot engine"):
//   - Slots start at each occurrence's start and follow each other every
//     duration + buffer time.
//   - A slot whose end falls after the occurrence's end is not offered.
//   - A slot overlapping an occupancy is not offered. Both sides count their
//     buffer: a candidate slot occupies [start, end + buffer) and must not
//     intersect [occupancy start, occupancy end).
//   - A slot starting before now + minimum notice, or after now + horizon, is
//     not offered.
//   - Overlapping windows offer each start time once.
func ComputeSlots(query SlotEngineQuery) []Slot {
	step := query.ProductDuration + query.BufferTime
	if step <= 0 || query.ProductDuration <= 0 {
		return nil
	}

	noticeCutoff := query.Now.Add(query.MinimumNotice)
	horizonCutoff := query.Now.Add(query.BookingHorizon)

	windows := make([]AvailabilityWindow, len(query.Availabilities))
	copy(windows, query.Availabilities)
	sortAvailabilityWindows(windows)

	var slots []Slot
	for _, window := range windows {
		// Slots are anchored to the occurrence's start; everything after is
		// duration arithmetic on instants, so no alignment rule applies here.
		// A window that began before the requested range jumps straight to
		// its first grid point inside it.
		start := window.Start
		if start.Before(query.From) {
			skipped := (query.From.Sub(start) + step - 1) / step
			start = start.Add(skipped * step)
		}
		for ; ; start = start.Add(step) {
			if !start.Before(query.To) {
				break // past the requested range (To exclusive)
			}
			end := start.Add(query.ProductDuration)
			if end.After(window.End) {
				break // this slot would end after the occurrence; later ones too
			}
			if start.Before(noticeCutoff) || start.After(horizonCutoff) {
				continue
			}
			if overlapsOccupancy(start, end.Add(query.BufferTime), query.Occupancies) {
				continue
			}
			slots = append(slots, Slot{AvailabilityID: window.AvailabilityID, Start: start, End: end})
		}
	}
	return orderedUniqueSlots(slots)
}

// orderedUniqueSlots orders slots by start and keeps one slot per start time:
// overlapping availabilities of the same partner (a one-off on top of a
// recurring series) may both offer a start, and the client picks a time, not
// an availability. The earliest-starting window's slot is kept.
func orderedUniqueSlots(slots []Slot) []Slot {
	sort.SliceStable(slots, func(i, j int) bool { return slots[i].Start.Before(slots[j].Start) })
	unique := slots[:0]
	for _, slot := range slots {
		if len(unique) > 0 && unique[len(unique)-1].Start.Equal(slot.Start) {
			continue
		}
		unique = append(unique, slot)
	}
	return unique
}

// overlapsOccupancy reports whether [slotStart, slotEnd) intersects any
// occupancy. Intervals are half-open: touching endpoints do not overlap.
func overlapsOccupancy(slotStart, slotEnd time.Time, occupancies []SlotOccupancy) bool {
	for _, occupancy := range occupancies {
		if slotStart.Before(occupancy.End) && occupancy.Start.Before(slotEnd) {
			return true
		}
	}
	return false
}

// sortAvailabilityWindows orders windows by start time.
func sortAvailabilityWindows(windows []AvailabilityWindow) {
	sort.SliceStable(windows, func(i, j int) bool { return windows[i].Start.Before(windows[j].Start) })
}

// OffersSlot reports whether the engine offers a slot starting exactly at
// start for the query: booking creation runs it so a start the page would not
// show is never accepted. The query's From/To are narrowed to that instant.
func OffersSlot(query SlotEngineQuery, start time.Time) bool {
	query.From = start
	query.To = start.Add(time.Nanosecond)
	for _, slot := range ComputeSlots(query) {
		if slot.Start.Equal(start) {
			return true
		}
	}
	return false
}
