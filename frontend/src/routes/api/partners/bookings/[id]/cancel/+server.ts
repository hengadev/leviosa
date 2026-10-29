import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// POST /api/partners/bookings/[id]/cancel — practitioner cancellation from the
// agenda: the client is refunded everything paid online and emailed. No body
// (no reason is collected, ADR-0019).
export const POST: RequestHandler = async ({ params, fetch }) => {
	const res = await fetch(`${env.API_URL}/partners/bookings/${params.id}/cancel`, {
		method: 'POST'
	});
	const data = await res.json().catch(() => ({}));
	return json(data, { status: res.status });
};
