// Calendar arithmetic for the booking card. Days are calendar dates in the
// practice's time zone (Europe/Paris), where the slot engine computes slots:
// a visitor in another time zone sees the practice's days and times.
//
// A day is carried as its key (YYYY-MM-DD) and as a Date at 12:00 UTC of that
// date, which formats to the same date in Europe/Paris (UTC+1/+2).

import { PRACTICE_TIME_ZONE } from "#lib/i18n/practice.ts";

export type CalendarMonth = { year: number; month: number }; // month: 0–11
export type CalendarDay = { key: string; date: Date };

const keyFormat = new Intl.DateTimeFormat("en-CA", { timeZone: PRACTICE_TIME_ZONE });

/** The practice-time-zone date (YYYY-MM-DD) of an instant. */
export const dayKey = (instant: Date): string => keyFormat.format(instant);

const pad = (n: number) => String(n).padStart(2, "0");

const keyOf = (year: number, month: number, day: number) => `${year}-${pad(month + 1)}-${pad(day)}`;

export const dayOf = (key: string): CalendarDay => ({ key, date: new Date(`${key}T12:00:00Z`) });

export function monthOf(key: string): CalendarMonth {
    const [year, month] = key.split("-").map(Number);
    return { year, month: month - 1 };
}

export function addMonths({ year, month }: CalendarMonth, count: number): CalendarMonth {
    const index = year * 12 + month + count;
    return { year: Math.floor(index / 12), month: index % 12 };
}

/** Negative when a is before b, 0 when the same month. */
export const compareMonths = (a: CalendarMonth, b: CalendarMonth) => a.year * 12 + a.month - (b.year * 12 + b.month);

/** A date inside the month, for formatting its name. */
export const monthDate = ({ year, month }: CalendarMonth) => new Date(Date.UTC(year, month, 15, 12));

/** The month as Monday-first weeks; null pads the first week. */
export function monthCells({ year, month }: CalendarMonth): (CalendarDay | null)[] {
    const padding = (new Date(Date.UTC(year, month, 1)).getUTCDay() + 6) % 7;
    const count = new Date(Date.UTC(year, month + 1, 0)).getUTCDate();
    const cells: (CalendarDay | null)[] = Array(padding).fill(null);
    for (let day = 1; day <= count; day++) cells.push(dayOf(keyOf(year, month, day)));
    return cells;
}
