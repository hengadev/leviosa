import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ fetch, url }) => {
	const origin = url.origin;
	try {
		const res = await fetch(`${env.API_URL}/users/me/stripe/billing-portal-link`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				return_url: `${origin}/client/profile?stripe=return`
			})
		});
		const data = await res.json();
		return json(data, { status: res.status });
	} catch {
		return json({ error: 'Failed to generate billing portal link' }, { status: 502 });
	}
};
