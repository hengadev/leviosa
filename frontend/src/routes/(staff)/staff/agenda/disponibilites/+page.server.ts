import type { PageServerLoad } from './$types';
import { redirect, isRedirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { practiceDate, type AvailabilityException } from './occurrences';

export interface Availability {
	id: string;
	/** Unique per occurrence: a recurring availability appears on several days. */
	key: string;
	date: string;
	startTime: string;
	endTime: string;
	status: 'available' | 'booked' | 'cancelled' | 'blocked';
	roomId: string;
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

export const load: PageServerLoad = async ({ locals, fetch }) => {
	if (!locals.user?.id) {
		throw redirect(302, '/auth');
	}

	const partnerId = locals.user.id;

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

		if (availRes.status === 401 || allocRes.status === 401) {
			throw redirect(302, '/auth');
		}

		// The recurring series (with their closed dates) and the one-off
		// availabilities; the calendar expands the series day by day.
		let series: Availability[] = [];
		let oneOffs: Availability[] = [];
		if (availRes.ok) {
			const data = await availRes.json();
			if (Array.isArray(data)) {
				const live = data.filter((a: any) => a.status !== 'cancelled');
				const recurring = live.filter((a: any) => a.recurrence_pattern);
				const exceptions = await loadExceptions(
					fetch,
					recurring.map((a: any) => a.id)
				);
				series = recurring.map((a: any) =>
					toAvailability(a, a.start_time, a.end_time, a.id, exceptions.get(a.id))
				);
				oneOffs = live
					.filter((a: any) => !a.recurrence_pattern)
					.map((a: any) => toAvailability(a, a.start_time, a.end_time, a.id));
			}
		} else {
			console.error(`Failed to fetch availabilities: ${availRes.status} ${availRes.statusText}`);
		}

		// The practitioner's room: a practitioner holds one, given when the
		// profile is created, so the form does not ask for it.
		let roomId: string | null = null;
		if (allocRes.ok) {
			const allocData = await allocRes.json();
			if (Array.isArray(allocData)) {
				roomId = allocData.find((a: any) => a.is_active)?.room_id ?? null;
			}
		} else {
			console.error(`Failed to fetch allocations: ${allocRes.status} ${allocRes.statusText}`);
		}

		return { series, oneOffs, roomId, partnerId };
	} catch (err) {
		if (isRedirect(err)) throw err;
		console.error('Error loading availabilities:', err);
		return { series: [], oneOffs: [], roomId: null, partnerId };
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
