import * as env from '$app/env/private';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// GET /api/notifications/push — whether practitioner notifications (Web Push)
// are available, with the VAPID public key the browser subscribes with.
export const GET: RequestHandler = async ({ fetch }) => {
	const res = await fetch(`${env.API_URL}/notifications/push`);
	const data = await res.json().catch(() => ({}));
	return json(data, { status: res.status });
};
