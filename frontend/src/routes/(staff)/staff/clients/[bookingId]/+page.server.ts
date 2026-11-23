import type { PageServerLoad } from './$types';
import { redirect, isRedirect } from '@sveltejs/kit';
import * as env from '$app/env/private';

export interface ClientSession {
	id: string;
	startTime: string;
	endTime: string;
	productName: string;
	status: 'confirmed' | 'completed' | 'no_show' | 'cancelled';
	paymentStatus: 'pending' | 'paid' | 'failed' | 'refunded' | 'partially_refunded' | 'disputed';
}

export interface ClientFile {
	name: string;
	phone: string;
	email: string;
	upcoming: ClientSession[];
	past: ClientSession[];
}

export const load: PageServerLoad = async ({ locals, params, fetch }) => {
	if (!locals.user?.id) {
		throw redirect(302, '/auth');
	}

	try {
		const res = await fetch(`${env.API_URL}/partners/bookings/${params.bookingId}/client`);

		if (res.status === 401) {
			throw redirect(302, '/auth');
		}

		// Not the caller's booking (or unknown): show the empty state.
		if (!res.ok) {
			return { client: null as ClientFile | null };
		}

		const data = await res.json();
		const toSession = (s: any): ClientSession => ({
			id: s.id,
			startTime: s.slot_start_time,
			endTime: s.slot_end_time,
			productName: s.product_name || 'Produit inconnu',
			status: s.status || 'confirmed',
			paymentStatus: s.payment_status || 'pending'
		});

		const client: ClientFile = {
			name: data.name || 'Client inconnu',
			phone: data.phone || '',
			email: data.email || '',
			upcoming: (data.upcoming || []).map(toSession),
			past: (data.past || []).map(toSession)
		};

		return { client };
	} catch (err) {
		if (isRedirect(err)) throw err;
		console.error('Error loading client file:', err);
		return { client: null as ClientFile | null };
	}
};
