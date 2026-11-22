package domain

import "time"

// Next working day (issue 037): the first Europe/Paris date after today on
// which the practitioner works, so Aujourd'hui can show it the evening
// before. A day counts as worked when it has at least one open availability
// occurrence (a recurring or one-off availability, not closed by an
// availability exception) or an appointment booking (confirmed, completed or
// no-show; never awaiting payment, expired or cancelled).

// NextWorkingDay returns the first Paris date strictly after today's date
// (in the practice time zone) with an open occurrence of one of the given
// availabilities or an appointment booking starting on it, looking ahead at
// most lookAheadDays days; ok is false when there is none. It is pure and
// reuses the occurrence expansion (Occurrences, WithoutExceptions).
func NextWorkingDay(availabilities []*Availability, exceptions []AvailabilityException, bookings []*Booking, today time.Time, lookAheadDays int) (time.Time, bool) {
	if lookAheadDays < 1 {
		return time.Time{}, false
	}

	first := civilDate(today.In(PracticeLocation)).AddDate(0, 0, 1)
	last := first.AddDate(0, 0, lookAheadDays - 1)
	from := onDate(first, midnight)
	to := onDate(last.AddDate(0, 0, 1), midnight)

	best := time.Time{}
	consider := func(day time.Time) {
		if !day.Before(first) && !day.After(last) && (best.IsZero() || day.Before(best)) {
			best = day
		}
	}

	for _, availability := range availabilities {
		if availability == nil || availability.Status != AvailabilityStatusAvailable {
			continue
		}
		for _, window := range availability.OccurrencesExcept(from, to, exceptions) {
			consider(civilDate(window.Start.In(PracticeLocation)))
		}
	}

	for _, booking := range bookings {
		if booking == nil || booking.Status == BookingStatusCancelled || !booking.IsAppointment() {
			continue
		}
		consider(civilDate(booking.SlotStartTime.In(PracticeLocation)))
	}

	if best.IsZero() {
		return time.Time{}, false
	}
	return onDate(best, midnight), true
}

// PartnerNextWorkingDayResponse is the partner endpoint's answer: the next
// working day's date as YYYY-MM-DD in the practice time zone (empty when
// there is none within the look-ahead) and that day's appointment bookings,
// in the same shape as the partner bookings list.
type PartnerNextWorkingDayResponse struct {
	Date     string                   `json:"date"`
	Bookings []PartnerBookingResponse `json:"bookings"`
}

// midnight is the wall-clock time of a date's start.
var midnight = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
