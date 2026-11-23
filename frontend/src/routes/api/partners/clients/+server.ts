import * as env from '$app/env/private';
import { error, json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// GET /api/partners/clients?partner_id=&q= — the logged-in practitioner's
// clients matching a name or phone fragment (issue 042's search), for the
// client step of the "Nouvelle réservation" form (issue 043). /api routes
// don't get locals.user (the hooks only resolve it on pages), so the page
// passes the practitioner's user ID; the backend checks the session is that
// partner or an administrator.
export const GET: RequestHandler = async ({ url, fetch }) => {
	const partnerId = url.searchParams.get('partner_id') ?? '';
	if (!/^[0-9a-f-]{36}$/i.test(partnerId)) throw error(400, 'Praticien inconnu');
	const query = new URLSearchParams();
	const q = url.searchParams.get('q');
	if (q) query.set('q', q);
	const res = await fetch(`${env.API_URL}/partners/clients/${partnerId}?${query}`);
	const data = await res.json().catch(() => ({}));
	return json(data, { status: res.status });
};
