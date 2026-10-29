import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// POST /api/bookings/[id]/cancel — cancel a booking as the signed-in client.
// No body is forwarded: the client cancellation collects no reason (ADR-0019).
export const POST: RequestHandler = async ({ params, fetch }) => {
	const res = await fetch(`${env.API_URL}/bookings/${params.id}/cancel`, {
		method: 'POST'
	});
	const data = await res.json();
	return json(data, { status: res.status });
};
