import * as env from '$app/env/private';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// POST /api/bookings/[id]/cancel-public — cancel a booking using a booking token (no auth required).
// No body: the client cancellation collects no reason (ADR-0019).
export const POST: RequestHandler = async ({ params, fetch, url }) => {
	const token = url.searchParams.get('token');
	if (!token) {
		return json({ error: 'Token manquant' }, { status: 400 });
	}

	const res = await fetch(`${env.API_URL}/bookings/${params.id}/cancel-public?token=${encodeURIComponent(token)}`, {
		method: 'POST'
	});
	const data = await res.json();
	return json(data, { status: res.status });
};
