// STATIC calendar data for the practice page's booking card. This is a
// placeholder that only demonstrates the layout: real free slots come from
// the public slot endpoint built on the slot engine in issue 010
// (docs/prd/single-practitioner-booking-mvp.md). Until then the calendar
// shows mock availability and must never be treated as bookable.

export type Slot = { id: string; start: Date; taken: boolean };
export type Day = { key: string; date: Date; slots: Slot[] };

export function dayKey(date: Date): string {
    const m = String(date.getMonth() + 1).padStart(2, "0");
    const d = String(date.getDate()).padStart(2, "0");
    return `${date.getFullYear()}-${m}-${d}`;
}

/** Mock availability: weekdays 9:00–18:00 (no 13:00), Saturday mornings, some slots taken. */
export function mockDays(count = 28, from = new Date()): Day[] {
    const days: Day[] = [];
    const start = new Date(from.getFullYear(), from.getMonth(), from.getDate());
    for (let i = 0; i < count; i++) {
        const date = new Date(start);
        date.setDate(start.getDate() + i);
        const weekday = date.getDay();
        const hours =
            weekday === 0 ? [] : weekday === 6 ? [9, 10, 11] : [9, 10, 11, 12, 14, 15, 16, 17, 18];
        const slots = hours
            .map((hour) => {
                const slotStart = new Date(date);
                slotStart.setHours(hour, 0, 0, 0);
                const seed = (date.getDate() * 31 + date.getMonth() * 7 + hour * 13) % 10;
                return { id: slotStart.toISOString(), start: slotStart, taken: seed < 4 };
            })
            .filter((slot) => slot.start.getTime() > from.getTime() + 60 * 60 * 1000);
        days.push({ key: dayKey(date), date, slots });
    }
    return days;
}

export const openSlots = (day: Day) => day.slots.filter((slot) => !slot.taken);
