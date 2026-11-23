import * as env from '$app/env/private';
import { error, json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// GET /api/partners/booking-slots?product_id=&from=YYYY-MM-DD&to=YYYY-MM-DD —
// the logged-in practitioner's free slots for the "Nouvelle réservation"
// form: no minimum notice and no booking horizon (issue 041). The practice
// page's public slots are a different endpoint. /api routes don't get
// locals.user (the hooks only resolve it on pages), so the page passes the
// practitioner's user ID; the backend checks the session is that partner or
// an administrator.
export const GET: RequestHandler = async ({ url, fetch }) => {
	const partnerId = url.searchParams.get('partner_id') ?? '';
	if (!/^[0-9a-f-]{36}$/i.test(partnerId)) throw error(400, 'Praticien inconnu');
	const query = new URLSearchParams();
	for (const key of ['product_id', 'from', 'to']) {
		const value = url.searchParams.get(key);
		if (value) query.set(key, value);
	}
	const res = await fetch(`${env.API_URL}/partners/booking-slots/${partnerId}?${query}`);
	const data = await res.json().catch(() => ({}));
	return json(data, { status: res.status });
};
