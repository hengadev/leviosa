import type { PageServerLoad } from './$types';
import { redirect, isRedirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { today } from '@internationalized/date';
import { PRACTICE_TIME_ZONE } from '$lib/i18n/practice';
import { occurrences, practiceDate, type AvailabilityException } from './occurrences';

export interface Availability {
	id: string;
	/** Unique per occurrence: a recurring availability appears on several days. */
	key: string;
	date: string;
	startTime: string;
	endTime: string;
	status: 'available' | 'booked' | 'cancelled' | 'blocked';
	roomId: string;
	roomName: string;
	maxCapacity: number;
	serviceType?: string;
	priceCents?: number;
	notes?: string;
	isRecurring: boolean;
	recurrencePattern?: {
		type: string;
		interval: number;
		until?: string;
		days_of_week?: number[];
	};
	/** A recurring series' availability exceptions (closed dates), by start date. */
	exceptions: AvailabilityException[];
	createdAt: string;
	updatedAt: string;
}

export interface RoomAllocation {
	id: string;
	roomId: string;
	allocationType: 'dedicated' | 'shared';
	startDate?: string;
	endDate?: string;
	isActive: boolean;
	roomName?: string;
}

export interface AvailabilityDay {
	date: string;
	dayName: string;
	slots: Availability[];
}

export const load: PageServerLoad = async ({ locals, url, fetch }) => {
	if (!locals.user?.id) {
		throw redirect(302, '/auth');
	}

	const partnerId = locals.user.id;
	const weekOffset = Math.max(0, Math.min(52, parseInt(url.searchParams.get('semaine') ?? '0', 10) || 0));

	try {
		// Fetch availabilities and allocations in parallel
		const [availRes, allocRes] = await Promise.all([
			fetch(`${env.API_URL}/partners/availabilities/${partnerId}`, {
				headers: { 'Content-Type': 'application/json' }
			}),
			fetch(`${env.API_URL}/partners/allocations/${partnerId}`, {
				headers: { 'Content-Type': 'application/json' }
			})
		]);

		if (availRes.status === 401) {
			throw redirect(302, '/auth');
		}

		// Process availabilities
		let availabilities: AvailabilityDay[] = emptyWeek(weekOffset);
		let series: Availability[] = [];
		if (availRes.ok) {
			const data = await availRes.json();
			if (Array.isArray(data)) {
				const recurring = data.filter((a: any) => a.recurrence_pattern && a.status !== 'cancelled');
				const exceptions = await loadExceptions(
					fetch,
					recurring.map((a: any) => a.id)
				);
				availabilities = groupByDay(data, weekOffset, exceptions);
				series = recurring.map((a: any) =>
					toAvailability(a, a.start_time, a.end_time, a.id, exceptions.get(a.id))
				);
			}
		} else {
			console.error(`Failed to fetch availabilities: ${availRes.status} ${availRes.statusText}`);
		}

		// Process allocations
		let allocations: RoomAllocation[] = [];
		if (allocRes.status === 401) {
			throw redirect(302, '/auth');
		}
		if (allocRes.ok) {
			const allocData = await allocRes.json();
			if (Array.isArray(allocData)) {
				allocations = allocData
					.filter((a: any) => a.is_active)
					.map((a: any): RoomAllocation => ({
						id: a.id,
						roomId: a.room_id,
						allocationType: a.allocation_type,
						startDate: a.start_date ?? undefined,
						endDate: a.end_date ?? undefined,
						isActive: a.is_active,
						roomName: a.room_name ?? undefined
					}));
			}
		} else {
			console.error(`Failed to fetch allocations: ${allocRes.status} ${allocRes.statusText}`);
		}

		return { availabilities, series, allocations, partnerId, weekOffset };
	} catch (err) {
		if (isRedirect(err)) throw err;
		console.error('Error loading availabilities:', err);
		return { availabilities: emptyWeek(weekOffset), series: [], allocations: [], partnerId, weekOffset };
	}
};

// The availability exceptions of each recurring series (issue 012). A series
// whose exceptions cannot be read shows all its occurrences.
async function loadExceptions(
	fetch: typeof globalThis.fetch,
	ids: string[]
): Promise<Map<string, AvailabilityException[]>> {
	const entries = await Promise.all(
		ids.map(async (id): Promise<[string, AvailabilityException[]]> => {
			try {
				const res = await fetch(`${env.API_URL}/availabilities/${id}/exceptions`, {
					headers: { 'Content-Type': 'application/json' }
				});
				if (!res.ok) {
					console.error(`Failed to fetch exceptions of ${id}: ${res.status} ${res.statusText}`);
					return [id, []];
				}
				const data = await res.json();
				return [id, Array.isArray(data) ? data.map(toException) : []];
			} catch (err) {
				console.error(`Failed to fetch exceptions of ${id}:`, err);
				return [id, []];
			}
		})
	);
	return new Map(entries);
}

function toException(e: any): AvailabilityException {
	return { id: e.id, startDate: e.start_date, endDate: e.end_date };
}

const dayNames = ['Dimanche', 'Lundi', 'Mardi', 'Mercredi', 'Jeudi', 'Vendredi', 'Samedi'];

function toAvailability(
	slot: any,
	startTime: string,
	endTime: string,
	key: string,
	exceptions: AvailabilityException[] = []
): Availability {
	return {
		id: slot.id,
		key,
		date: practiceDate(startTime).toString(),
		startTime,
		endTime,
		status: slot.status || 'available',
		roomId: slot.room_id,
		roomName: slot.room_id || 'Salle inconnue',
		maxCapacity: slot.max_capacity || 1,
		serviceType: slot.service_type,
		priceCents: slot.price_cents ?? undefined,
		notes: slot.notes,
		isRecurring: !!slot.recurrence_pattern,
		recurrencePattern: slot.recurrence_pattern ?? undefined,
		exceptions,
		createdAt: slot.created_at,
		updatedAt: slot.updated_at
	};
}

// The week shown: seven days from today (practice time zone), weekOffset weeks ahead.
function weekDays(weekOffset: number) {
	const first = today(PRACTICE_TIME_ZONE).add({ weeks: weekOffset });
	return Array.from({ length: 7 }, (_, i) => first.add({ days: i }));
}

function emptyWeek(weekOffset: number): AvailabilityDay[] {
	return groupByDay([], weekOffset, new Map());
}

// One-off availabilities on their day, and every occurrence of the recurring
// ones that fall in the week (a deleted series shows no occurrence, and a
// closed date none of its series' occurrences).
function groupByDay(
	slots: any[],
	weekOffset: number,
	exceptions: Map<string, AvailabilityException[]>
): AvailabilityDay[] {
	const days = weekDays(weekOffset);
	const first = days[0];
	const last = days[days.length - 1];
	const slotsByDate = new Map<string, Availability[]>();
	const add = (availability: Availability) => {
		if (!slotsByDate.has(availability.date)) slotsByDate.set(availability.date, []);
		slotsByDate.get(availability.date)!.push(availability);
	};

	for (const slot of slots) {
		if (slot.recurrence_pattern) {
			if (slot.status === 'cancelled') continue;
			const closed = exceptions.get(slot.id) ?? [];
			for (const o of occurrences(slot.start_time, slot.end_time, slot.recurrence_pattern, first, last, closed)) {
				add(toAvailability(slot, o.start, o.end, `${slot.id}-${o.start}`));
			}
		} else {
			add(toAvailability(slot, slot.start_time, slot.end_time, slot.id));
		}
	}

	const availabilities: AvailabilityDay[] = [];
	for (const d of days) {
		const dateKey = d.toString();
		const daySlots = (slotsByDate.get(dateKey) || []).sort((a, b) =>
			a.startTime.localeCompare(b.startTime)
		);
		const weekdayIndex = new Date(Date.UTC(d.year, d.month - 1, d.day)).getUTCDay();
		if (daySlots.length > 0 || weekdayIndex !== 0) {
			availabilities.push({ date: dateKey, dayName: dayNames[weekdayIndex], slots: daySlots });
		}
	}
	return availabilities;
}
