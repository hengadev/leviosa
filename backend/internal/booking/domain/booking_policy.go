package domain

import "time"

// BookingPolicy holds the practice-wide limits on when a client may book: no
// slot closer than MinimumNotice to now, none further than Horizon from now.
// Its values come from the settings (booking policy), so an administrator
// changes them without a release.
type BookingPolicy struct {
	MinimumNotice time.Duration
	Horizon       time.Duration
}

const (
	// DefaultMinimumNotice applies while the setting is unset (PRD "Settings").
	DefaultMinimumNotice = 12 * time.Hour
	// DefaultBookingHorizon applies while the setting is unset: 8 weeks.
	DefaultBookingHorizon = 8 * 7 * 24 * time.Hour
)

// DefaultBookingPolicy is the policy used when none is configured.
func DefaultBookingPolicy() BookingPolicy {
	return BookingPolicy{MinimumNotice: DefaultMinimumNotice, Horizon: DefaultBookingHorizon}
}

// OccupiesTime reports whether the booking blocks its slot for other
// clients. Every booking except a cancelled one does; issue 014 adds
// `awaiting_payment` (while its hold runs) and `expired` (never).
func (b *Booking) OccupiesTime() bool {
	return b.Status != BookingStatusCancelled
}

// Occupancy is the interval the booking takes in the engine's eyes: its
// session plus the booked product's buffer time.
func (b *Booking) Occupancy(bufferTime time.Duration) SlotOccupancy {
	return SlotOccupancy{Start: b.SlotStartTime, End: b.SlotEndTime.Add(bufferTime)}
}
