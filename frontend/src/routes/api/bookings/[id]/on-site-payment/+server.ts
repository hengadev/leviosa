import * as env from '$app/env/private';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

async function forward(id: string, method: 'POST' | 'DELETE', fetch: typeof globalThis.fetch) {
	const res = await fetch(`${env.API_URL}/bookings/${id}/on-site-payment`, {
		method,
		headers: { 'Content-Type': 'application/json' }
	});
	const data = await res.json().catch(() => ({}));
	return json(data, { status: res.status });
}

// POST /api/bookings/[id]/on-site-payment — record that the amount due at the practice was paid
export const POST: RequestHandler = async ({ params, fetch }) => forward(params.id, 'POST', fetch);

// DELETE /api/bookings/[id]/on-site-payment — undo it, the same Paris day only
export const DELETE: RequestHandler = async ({ params, fetch }) =>
	forward(params.id, 'DELETE', fetch);
