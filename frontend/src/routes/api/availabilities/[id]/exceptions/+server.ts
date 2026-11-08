import * as env from '$app/env/private';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// GET /api/availabilities/[id]/exceptions — closed dates of a recurring availability
export const GET: RequestHandler = async ({ params, fetch }) => {
	const res = await fetch(`${env.API_URL}/availabilities/${params.id}/exceptions`, {
		headers: { 'Content-Type': 'application/json' }
	});
	const data = await res.json();
	return json(data, { status: res.status });
};

// POST /api/availabilities/[id]/exceptions — close a date or a date range
export const POST: RequestHandler = async ({ params, request, fetch }) => {
	const body = await request.json();
	const res = await fetch(`${env.API_URL}/availabilities/${params.id}/exceptions`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
	const data = await res.json();
	return json(data, { status: res.status });
};
