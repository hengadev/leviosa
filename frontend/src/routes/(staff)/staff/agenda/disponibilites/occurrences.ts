import {
	CalendarDate,
	CalendarDateTime,
	parseAbsolute,
	parseDateTime,
	toCalendarDate,
	toZoned,
	type ZonedDateTime
} from '@internationalized/date';
import { PRACTICE_TIME_ZONE } from '$lib/i18n/practice';

// Occurrences of a recurring availability, for the partner agenda. Mirrors
// the backend's expansion (backend/internal/booking/domain/recurrence.go),
// which is what the practice page offers: wall-clock times in Europe/Paris,
// the series starts on its first day, weekly weeks start on Monday, a monthly
// series skips months without its day, and the end date is inclusive.

export interface RecurrencePattern {
	type: string;
	interval: number;
	until?: string;
	days_of_week?: number[];
}

export interface OccurrenceWindow {
	start: string; // ISO instant
	end: string;
}

const zone = PRACTICE_TIME_ZONE;

/** The calendar date of an instant in the practice time zone. */
export function practiceDate(iso: string): CalendarDate {
	return toCalendarDate(parseAbsolute(iso, zone));
}

/** An ISO instant for a wall-clock date (YYYY-MM-DD) and time (HH:MM) in the practice time zone. */
export function practiceInstant(date: string, time: string): string {
	return toZoned(parseDateTime(`${date}T${time}`), zone).toAbsoluteString();
}

function daysBetween(a: CalendarDate, b: CalendarDate): number {
	return b.compare(a);
}

// 0 = Sunday … 6 = Saturday, like the backend's days_of_week.
function weekday(d: CalendarDate): number {
	return new Date(Date.UTC(d.year, d.month - 1, d.day)).getUTCDay();
}

function mondayOf(d: CalendarDate): CalendarDate {
	return d.subtract({ days: (weekday(d) + 6) % 7 });
}

function occursOn(p: RecurrencePattern, day: CalendarDate, anchor: CalendarDate): boolean {
	const interval = Math.max(1, p.interval || 1);
	switch (p.type) {
		case 'daily':
			return daysBetween(anchor, day) % interval === 0;
		case 'weekly': {
			const days = p.days_of_week?.length ? p.days_of_week : [weekday(anchor)];
			if (!days.includes(weekday(day))) return false;
			const weeks = daysBetween(mondayOf(anchor), mondayOf(day)) / 7;
			return weeks % interval === 0;
		}
		case 'monthly': {
			if (day.day !== anchor.day) return false;
			const months = (day.year - anchor.year) * 12 + day.month - anchor.month;
			return months % interval === 0;
		}
		default:
			return day.compare(anchor) === 0;
	}
}

function at(day: CalendarDate, clock: ZonedDateTime): string {
	return toZoned(
		new CalendarDateTime(day.year, day.month, day.day, clock.hour, clock.minute),
		zone
	).toAbsoluteString();
}

/**
 * The occurrences of a series starting on the calendar days [first, last]
 * (practice time zone), given its first occurrence (startIso, endIso).
 */
export function occurrences(
	startIso: string,
	endIso: string,
	pattern: RecurrencePattern,
	first: CalendarDate,
	last: CalendarDate
): OccurrenceWindow[] {
	const anchorStart = parseAbsolute(startIso, zone);
	const anchorEnd = parseAbsolute(endIso, zone);
	const anchor = toCalendarDate(anchorStart);
	const span = daysBetween(anchor, toCalendarDate(anchorEnd));

	const from = first.compare(anchor) < 0 ? anchor : first;
	let to = last;
	if (pattern.until) {
		const until = practiceDate(pattern.until);
		if (until.compare(to) < 0) to = until;
	}

	const out: OccurrenceWindow[] = [];
	for (let day = from; day.compare(to) <= 0; day = day.add({ days: 1 })) {
		if (!occursOn(pattern, day, anchor)) continue;
		out.push({ start: at(day, anchorStart), end: at(day.add({ days: span }), anchorEnd) });
	}
	return out;
}
