import type { PageServerLoad } from './$types';
import { redirect, isRedirect } from '@sveltejs/kit';
import * as env from '$app/env/private';

export interface Booking {
	id: string;
	clientId: string;
	clientName: string;
	productName: string;
	startTime: string;
	endTime: string;
	status: 'confirmed' | 'completed' | 'no_show' | 'cancelled';
	paymentStatus: 'pending' | 'paid' | 'failed' | 'refunded' | 'partially_refunded' | 'disputed';
	amountInCents: number;
	currency: string;
	/** What cancelling from the agenda refunds the client: all they paid online. */
	practitionerRefundCents: number;
	/** When the booking was closed, if it was (completed or no-show). */
	completedAt?: string;
	noShowAt?: string;
}

export const load: PageServerLoad = async ({ locals, url, fetch, parent }) => {
	if (!locals.user?.id) {
		throw redirect(302, '/auth');
	}

	const partnerId = locals.user.id;
	const statusFilter = url.searchParams.get('status') ?? '';

	// An administrator without a practitioner profile of their own has no
	// bookings to list: the page says the profile is not set up.
	if (locals.user.role === 'administrator' && !(await parent()).selfPractitioner) {
		return { bookings: [] as Booking[], statusFilter, profileMissing: true };
	}

	try {
		const res = await fetch(`${env.API_URL}/partners/bookings/${partnerId}`);

		if (res.status === 401) {
			throw redirect(302, '/auth');
		}

		if (!res.ok) {
			console.error(`Failed to fetch bookings: ${res.status} ${res.statusText}`);
			return { bookings: [], statusFilter };
		}

		const data = await res.json();

		if (!Array.isArray(data)) {
			console.error('Invalid response format: expected array');
			return { bookings: [], statusFilter };
		}

		const bookings: Booking[] = data.map((booking: any) => {
			return {
				id: booking.id,
				clientId: booking.client_id,
				clientName: booking.client_name || 'Client inconnu',
				productName: booking.product_name || 'Produit inconnu',
				startTime: booking.slot_start_time,
				endTime: booking.slot_end_time,
				status: booking.status || 'confirmed',
				paymentStatus: booking.payment_status || 'pending',
				amountInCents: booking.total_price_cents || 0,
				currency: booking.currency || 'EUR',
				practitionerRefundCents: booking.practitioner_refund_cents || 0,
				completedAt: booking.completed_at ?? undefined,
				noShowAt: booking.no_show_at ?? undefined
			};
		});

		bookings.sort((a, b) => new Date(a.startTime).getTime() - new Date(b.startTime).getTime());

		return { bookings, statusFilter };
	} catch (err) {
		if (isRedirect(err)) throw err;
		console.error('Error loading bookings:', err);
		return { bookings: [], statusFilter };
	}
};
