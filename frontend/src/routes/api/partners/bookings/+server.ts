import * as env from '$app/env/private';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// POST /api/partners/bookings — the practitioner books a client who called
// (issue 041): confirmed at once, paid at the practice, and the client gets
// no message from the app. 409 when the slot was taken meanwhile.
export const POST: RequestHandler = async ({ request, fetch }) => {
	const body = await request.text();
	const res = await fetch(`${env.API_URL}/partners/bookings`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body
	});
	const data = await res.json().catch(() => ({}));
	return json(data, { status: res.status });
};
