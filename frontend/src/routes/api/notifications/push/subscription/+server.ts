import * as env from '$app/env/private';
import { error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// The current device's push subscription (the Cabinet switch, ADR-0020).
// PUT stores it for the signed-in practitioner, DELETE removes it.

async function forward(method: 'PUT' | 'DELETE', body: unknown, fetchFn: typeof fetch) {
	const res = await fetchFn(`${env.API_URL}/notifications/push/subscription`, {
		method,
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
	if (res.status === 204) return new Response(null, { status: 204 });
	const text = await res.text();
	return new Response(text, {
		status: res.status,
		headers: { 'Content-Type': res.headers.get('Content-Type') ?? 'application/json' }
	});
}

export const PUT: RequestHandler = async ({ request, fetch }) => {
	const body = await request.json().catch(() => null);
	if (!body || typeof body !== 'object') error(400, 'invalid subscription');
	const { endpoint, keys } = body as { endpoint?: unknown; keys?: unknown };
	return forward(
		'PUT',
		{ endpoint, keys, user_agent: request.headers.get('user-agent') ?? '' },
		fetch
	);
};

export const DELETE: RequestHandler = async ({ request, fetch }) => {
	const body = await request.json().catch(() => null);
	const endpoint = body && typeof body === 'object' ? (body as { endpoint?: unknown }).endpoint : undefined;
	if (typeof endpoint !== 'string' || !endpoint) error(400, 'endpoint is required');
	return forward('DELETE', { endpoint }, fetch);
};
