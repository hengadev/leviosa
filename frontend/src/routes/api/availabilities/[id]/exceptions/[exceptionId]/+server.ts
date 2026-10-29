import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// DELETE /api/availabilities/[id]/exceptions/[exceptionId] — reopen closed dates
export const DELETE: RequestHandler = async ({ params, fetch }) => {
	const res = await fetch(
		`${env.API_URL}/availabilities/${params.id}/exceptions/${params.exceptionId}`,
		{ method: 'DELETE', headers: { 'Content-Type': 'application/json' } }
	);
	if (res.status === 204) return new Response(null, { status: 204 });
	const data = await res.json();
	return json(data, { status: res.status });
};
