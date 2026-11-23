package domain

import "time"

// BookingPolicy holds the practice-wide limits on when a client may book: no
// slot closer than MinimumNotice to now, none further than Horizon from now.
// PaymentHold is how long a new, unpaid booking keeps its slot while the
// client pays (ADR-0018). Its values come from the settings (booking
// policy), so an administrator changes them without a release.
type BookingPolicy struct {
	MinimumNotice time.Duration
	Horizon       time.Duration
	PaymentHold   time.Duration
}

const (
	// DefaultMinimumNotice applies while the setting is unset (PRD "Settings").
	DefaultMinimumNotice = 12 * time.Hour
	// DefaultBookingHorizon applies while the setting is unset: 8 weeks.
	DefaultBookingHorizon = 8 * 7 * 24 * time.Hour
	// DefaultPaymentHold applies while the setting is unset (ADR-0018).
	DefaultPaymentHold = 15 * time.Minute
)

// PractitionerBookingHorizon stands for "no horizon" on the practitioner's
// bookings (ADR-0021): far enough that no agenda reaches it.
const PractitionerBookingHorizon = 10 * 365 * 24 * time.Hour

// ForPractitioner is the policy the practitioner books under for a client
// who called (ADR-0021): neither the minimum notice nor the booking horizon
// applies, only the availabilities and the bookings do. A slot that has
// already started is still not offered.
func (p BookingPolicy) ForPractitioner() BookingPolicy {
	p.MinimumNotice = 0
	p.Horizon = PractitionerBookingHorizon
	return p
}

// DefaultBookingPolicy is the policy used when none is configured.
func DefaultBookingPolicy() BookingPolicy {
	return BookingPolicy{MinimumNotice: DefaultMinimumNotice, Horizon: DefaultBookingHorizon, PaymentHold: DefaultPaymentHold}
}

// OccupiesTime reports whether the booking blocks its slot for other
// clients at now. A cancelled or expired booking never does; an
// `awaiting_payment` booking does only while its payment hold runs
// (ADR-0018), so an elapsed hold frees the slot even before the expiry job
// moves the booking to `expired`. Every other status does.
func (b *Booking) OccupiesTime(now time.Time) bool {
	switch b.Status {
	case BookingStatusCancelled, BookingStatusExpired:
		return false
	case BookingStatusAwaitingPayment:
		// A hold without an expiry cannot be released by time: keep the slot
		// rather than risk two clients paying for it.
		return b.HoldExpiresAt == nil || now.Before(*b.HoldExpiresAt)
	default:
		return true
	}
}

// Occupancy is the interval the booking takes in the engine's eyes: its
// session plus the booked product's buffer time.
func (b *Booking) Occupancy(bufferTime time.Duration) SlotOccupancy {
	return SlotOccupancy{Start: b.SlotStartTime, End: b.SlotEndTime.Add(bufferTime)}
}

// HoldElapsed reports whether the booking awaits payment past its payment
// hold at now: it no longer keeps its slot, even before the expiry job moves
// it to expired. A hold without an expiry never elapses.
func (b *Booking) HoldElapsed(now time.Time) bool {
	return b.Status == BookingStatusAwaitingPayment && b.HoldExpiresAt != nil && !now.Before(*b.HoldExpiresAt)
}

// OverlapsAny reports whether the booking, extended by its product's buffer
// time, intersects any of the occupancies: whether its slot is still free for
// it, judged as the slot engine judges a candidate slot. Callers leave the
// booking's own occupancy out.
func (b *Booking) OverlapsAny(bufferTime time.Duration, occupancies []SlotOccupancy) bool {
	return overlapsOccupancy(b.SlotStartTime, b.SlotEndTime.Add(bufferTime), occupancies)
}
