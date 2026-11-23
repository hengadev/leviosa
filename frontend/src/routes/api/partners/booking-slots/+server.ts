import * as env from '$app/env/private';
import { error, json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// GET /api/partners/booking-slots?product_id=&from=YYYY-MM-DD&to=YYYY-MM-DD —
// the logged-in practitioner's free slots for the "Nouvelle réservation"
// form: no minimum notice and no booking horizon (issue 041). The practice
// page's public slots are a different endpoint.
export const GET: RequestHandler = async ({ locals, url, fetch }) => {
	if (!locals.user?.id) throw error(401, 'Non connecté');
	const query = new URLSearchParams();
	for (const key of ['product_id', 'from', 'to']) {
		const value = url.searchParams.get(key);
		if (value) query.set(key, value);
	}
	const res = await fetch(`${env.API_URL}/partners/booking-slots/${locals.user.id}?${query}`);
	const data = await res.json().catch(() => ({}));
	return json(data, { status: res.status });
};
