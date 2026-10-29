package domain

import "time"

// Recurrence expansion (docs/issues/011-recurring-availabilities.md).
//
// A recurring availability is stored once: its StartTime/EndTime are the
// series' first occurrence (the anchor) and its RecurrencePattern says how it
// repeats. The slot engine only ever sees occurrence windows, so callers turn
// a series into the windows of a range with Occurrences before building the
// engine's query.
//
// Rules, all evaluated on calendar dates in the practice time zone
// (PracticeLocation, Europe/Paris):
//
//   - Every occurrence starts at the anchor's wall-clock start time and ends
//     at the anchor's wall-clock end time (on the same day, or as many days
//     later as the anchor spans). A weekly 09:00–13:00 therefore stays
//     09:00–13:00 across a daylight-saving change: the instants shift, the
//     wall clock does not.
//   - The series starts on the anchor's date: no occurrence falls before it.
//   - daily: every Interval days from the anchor's date.
//   - weekly: on each listed day of the week (DaysOfWeek, 0 = Sunday), in
//     every Interval-th week counted from the anchor's week. Weeks start on
//     Monday (the French calendar). The anchor's date is an occurrence only
//     when its day of the week is listed. A weekly pattern without days
//     (never accepted at creation) repeats on the anchor's day of the week.
//   - monthly: on the anchor's day of the month, every Interval months from
//     the anchor's month. A month without that day (the 31st in April, the
//     30th in February) has no occurrence: the series never drifts to
//     another day (RFC 5545 behaviour).
//   - Until, when set, is the series' last day, inclusive: an occurrence on
//     Until's date (in the practice time zone) is still offered, whatever
//     time of day Until holds.
//   - An unknown pattern type yields the anchor occurrence only.
//
// A one-off availability is its own single occurrence.

// Recurrence pattern types.
const (
	RecurrenceDaily   = "daily"
	RecurrenceWeekly  = "weekly"
	RecurrenceMonthly = "monthly"
)

// Occurrences returns the availability's occurrence windows that overlap
// [from, to), ordered by start. It is pure: callers pass the range, and it
// does not look at the availability's status.
func (a *Availability) Occurrences(from, to time.Time) []AvailabilityWindow {
	if !from.Before(to) {
		return nil
	}

	anchor := AvailabilityWindow{AvailabilityID: a.ID, Start: a.StartTime, End: a.EndTime}
	if !a.IsRecurring || a.RecurrencePattern == nil {
		if overlaps(anchor, from, to) {
			return []AvailabilityWindow{anchor}
		}
		return nil
	}

	pattern := *a.RecurrencePattern
	interval := pattern.Interval
	if interval < 1 {
		interval = 1
	}

	anchorStart := a.StartTime.In(PracticeLocation)
	anchorEnd := a.EndTime.In(PracticeLocation)
	anchorDate := civilDate(anchorStart)
	spanDays := daysBetween(anchorDate, civilDate(anchorEnd))

	// An occurrence starting up to spanDays (+1 for daylight-saving slack)
	// before from may still be running at from.
	first := civilDate(from.In(PracticeLocation)).AddDate(0, 0, -spanDays-1)
	if first.Before(anchorDate) {
		first = anchorDate
	}
	last := civilDate(to.In(PracticeLocation))
	if pattern.Until != nil {
		if until := civilDate(pattern.Until.In(PracticeLocation)); until.Before(last) {
			last = until
		}
	}

	var windows []AvailabilityWindow
	for day := first; !day.After(last); day = day.AddDate(0, 0, 1) {
		if !pattern.occursOn(day, anchorDate, interval) {
			continue
		}
		endDay := day.AddDate(0, 0, spanDays)
		window := AvailabilityWindow{
			AvailabilityID: a.ID,
			Start:          onDate(day, anchorStart),
			End:            onDate(endDay, anchorEnd),
		}
		if !window.End.After(window.Start) {
			continue // a daylight-saving gap swallowed the occurrence
		}
		if overlaps(window, from, to) {
			windows = append(windows, window)
		}
	}
	return windows
}

// SeriesEndedBefore reports whether a recurring availability has no
// occurrence on or after the given instant's date (its Until is past). A
// one-off availability has ended once its end is not after the instant.
func (a *Availability) SeriesEndedBefore(instant time.Time) bool {
	if !a.IsRecurring || a.RecurrencePattern == nil {
		return !a.EndTime.After(instant)
	}
	until := a.RecurrencePattern.Until
	if until == nil {
		return false
	}
	return civilDate(until.In(PracticeLocation)).Before(civilDate(instant.In(PracticeLocation)))
}

// occursOn reports whether the series has an occurrence starting on day
// (a civil date on or after the anchor's date).
func (p RecurrencePattern) occursOn(day, anchorDate time.Time, interval int) bool {
	switch p.Type {
	case RecurrenceDaily:
		return daysBetween(anchorDate, day)%interval == 0
	case RecurrenceWeekly:
		if !p.listsWeekday(day.Weekday(), anchorDate.Weekday()) {
			return false
		}
		weeks := daysBetween(mondayOf(anchorDate), mondayOf(day)) / 7
		return weeks%interval == 0
	case RecurrenceMonthly:
		if day.Day() != anchorDate.Day() {
			return false
		}
		months := (day.Year()-anchorDate.Year())*12 + int(day.Month()) - int(anchorDate.Month())
		return months%interval == 0
	default:
		return day.Equal(anchorDate)
	}
}

func (p RecurrencePattern) listsWeekday(weekday, anchorWeekday time.Weekday) bool {
	if len(p.DaysOfWeek) == 0 {
		return weekday == anchorWeekday
	}
	for _, listed := range p.DaysOfWeek {
		if time.Weekday(listed) == weekday {
			return true
		}
	}
	return false
}

func overlaps(window AvailabilityWindow, from, to time.Time) bool {
	return window.Start.Before(to) && window.End.After(from)
}

// civilDate is the calendar date of t (in t's location) as midnight UTC, so
// date arithmetic never meets a daylight-saving change.
func civilDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func daysBetween(from, to time.Time) int {
	return int(to.Sub(from).Hours() / 24)
}

// mondayOf returns the Monday of the week containing the civil date day.
func mondayOf(day time.Time) time.Time {
	offset := (int(day.Weekday()) + 6) % 7 // Monday 0 … Sunday 6
	return day.AddDate(0, 0, -offset)
}

// onDate returns the instant at clock's wall-clock time on the civil date
// day, in the practice time zone.
func onDate(day, clock time.Time) time.Time {
	return time.Date(day.Year(), day.Month(), day.Day(),
		clock.Hour(), clock.Minute(), clock.Second(), clock.Nanosecond(), PracticeLocation)
}
