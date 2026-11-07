import type { CalendarDate } from '@internationalized/date';
import type { Availability } from './+page.server';
import { isClosed, occurrences, practiceDate, type AvailabilityException } from './occurrences';

// What the availability calendar shows of each day: the openings of the
// recurring series (closed or not) and the one-off availabilities.

/** 0 = Sunday … 6 = Saturday, like the backend's days_of_week. */
export function weekday(d: CalendarDate): number {
	return new Date(Date.UTC(d.year, d.month - 1, d.day)).getUTCDay();
}

/** The weeks (Monday first) covering the month of `month`. */
export function monthGrid(month: CalendarDate): CalendarDate[][] {
	const first = month.set({ day: 1 });
	const last = first.add({ months: 1 }).subtract({ days: 1 });
	let day = first.subtract({ days: (weekday(first) + 6) % 7 });
	const weeks: CalendarDate[][] = [];
	while (day.compare(last) <= 0) {
		const week: CalendarDate[] = [];
		for (let i = 0; i < 7; i++) {
			week.push(day);
			day = day.add({ days: 1 });
		}
		weeks.push(week);
	}
	return weeks;
}

export interface Opening {
	availability: Availability;
	/** ISO instants of this day's opening. */
	start: string;
	end: string;
	/** The exception closing a series on this day, if any. */
	closedBy?: AvailabilityException;
}

/** The openings of a day, by start time. */
export function dayOpenings(day: CalendarDate, series: Availability[], oneOffs: Availability[]): Opening[] {
	const date = day.toString();
	const out: Opening[] = [];
	for (const s of series) {
		if (!s.recurrencePattern) continue;
		for (const o of occurrences(s.startTime, s.endTime, s.recurrencePattern, day, day)) {
			const closedBy = isClosed(date, s.exceptions)
				? s.exceptions.find((e) => e.startDate <= date && date <= e.endDate)
				: undefined;
			out.push({ availability: s, start: o.start, end: o.end, closedBy });
		}
	}
	for (const a of oneOffs) {
		if (practiceDate(a.startTime).toString() === date) {
			out.push({ availability: a, start: a.startTime, end: a.endTime });
		}
	}
	return out.sort((a, b) => a.start.localeCompare(b.start));
}

/** The first day from `from` (inclusive) on one of the weekdays, up to `until` if set. */
export function firstWeekday(from: CalendarDate, weekdays: number[], until?: CalendarDate): CalendarDate | null {
	if (weekdays.length === 0) return null;
	for (let d = from, i = 0; i < 7; d = d.add({ days: 1 }), i++) {
		if (until && d.compare(until) > 0) return null;
		if (weekdays.includes(weekday(d))) return d;
	}
	return null;
}

/**
 * The open opening (not closed by an exception) that [start, end) overlaps
 * on its day, if any. The backend's conflict check compares with a series'
 * first occurrence only, so the agenda checks every occurrence itself.
 */
export function overlapping(
	day: CalendarDate,
	start: string,
	end: string,
	series: Availability[],
	oneOffs: Availability[],
	/** An availability left out (one whose day is being replaced). */
	ignoreId?: string
): Opening | undefined {
	// Instants are compared parsed: the backend and the calendar write their offsets differently.
	const [s, e] = [Date.parse(start), Date.parse(end)];
	return dayOpenings(day, series, oneOffs).find(
		(o) =>
			!o.closedBy && o.availability.id !== ignoreId && Date.parse(o.start) < e && s < Date.parse(o.end)
	);
}

/** The first day of a weekly availability that overlaps an open opening. */
export function firstWeeklyClash(
	first: CalendarDate,
	until: CalendarDate | undefined,
	weekdays: number[],
	at: (day: CalendarDate) => { start: string; end: string },
	series: Availability[],
	oneOffs: Availability[]
): CalendarDate | undefined {
	// Without an end date, a year ahead is checked: further than anyone books.
	const last = until ?? first.add({ years: 1 });
	for (let d = first; d.compare(last) <= 0; d = d.add({ days: 1 })) {
		if (!weekdays.includes(weekday(d))) continue;
		const { start, end } = at(d);
		if (overlapping(d, start, end, series, oneOffs)) return d;
	}
	return undefined;
}
