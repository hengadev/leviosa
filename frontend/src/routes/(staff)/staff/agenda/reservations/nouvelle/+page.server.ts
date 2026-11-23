import type { PageServerLoad } from './$types';
import * as env from '$app/env/private';
import { redirect, isRedirect } from '@sveltejs/kit';

// "Nouvelle réservation" (issue 041): the practitioner books a client who
// called, from the "⋮" menu of Aujourd'hui or Réservations. The booking is
// confirmed at once, the full price is due at the practice, and the app
// sends the client nothing.

/** A product the practitioner can book (GET /partners/booking-products). */
export interface BookableProduct {
	id: string;
	name: string;
	durationMinutes: number;
	priceCents: number;
	currency: string;
}

/** The pages the form can be opened from, and returns to. */
const RETURN_PAGES = ['/admin', '/staff/agenda/reservations'] as const;
export type ReturnPage = (typeof RETURN_PAGES)[number];

function returnPage(from: string | null): ReturnPage {
	return RETURN_PAGES.find((page) => page === from) ?? '/staff/agenda/reservations';
}

export const load: PageServerLoad = async ({ locals, url, fetch, parent }) => {
	if (!locals.user?.id) {
		throw redirect(302, '/auth');
	}
	const from = returnPage(url.searchParams.get('from'));

	// An administrator without a practitioner profile of their own has no
	// agenda to book in.
	if (locals.user.role === 'administrator' && !(await parent()).selfPractitioner) {
		return { from, products: [] as BookableProduct[], profileMissing: true, loadFailed: false };
	}
	if (env.USE_MOCK_DATA === 'true') {
		return { from, products: [] as BookableProduct[], profileMissing: false, loadFailed: false };
	}

	try {
		const res = await fetch(`${env.API_URL}/partners/booking-products/${locals.user.id}`);
		if (res.status === 401) {
			throw redirect(302, '/auth');
		}
		if (!res.ok) {
			console.error(`Failed to fetch the bookable products: ${res.status} ${res.statusText}`);
			return { from, products: [] as BookableProduct[], profileMissing: false, loadFailed: true };
		}
		const data = await res.json();
		const products: BookableProduct[] = (Array.isArray(data) ? data : []).map((p: any) => ({
			id: p.id,
			name: p.name || 'Soin',
			durationMinutes: p.duration_minutes ?? 0,
			priceCents: p.price_cents ?? 0,
			currency: p.currency || 'EUR'
		}));
		return { from, products, profileMissing: false, loadFailed: false };
	} catch (err) {
		if (isRedirect(err)) throw err;
		console.error('Error loading the bookable products:', err);
		return { from, products: [] as BookableProduct[], profileMissing: false, loadFailed: true };
	}
};
