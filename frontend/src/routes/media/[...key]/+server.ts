import { env } from '$env/dynamic/private';
import type { RequestHandler } from './$types';

// Uploaded media (the practice page video) is served by the backend, which
// streams it from the private media bucket. This relays it unchanged, byte
// ranges included, so the page can use a same-origin /media/... URL in every
// environment. The backend decides which keys exist.
const FORWARDED_REQUEST_HEADERS = ['range', 'if-none-match'];
const FORWARDED_RESPONSE_HEADERS = [
	'accept-ranges',
	'cache-control',
	'content-length',
	'content-range',
	'content-type',
	'etag',
	'last-modified',
	'x-content-type-options'
];

const relay: RequestHandler = async ({ params, request }) => {
	const headers = new Headers();
	for (const name of FORWARDED_REQUEST_HEADERS) {
		const value = request.headers.get(name);
		if (value) headers.set(name, value);
	}

	let res: Response;
	try {
		res = await fetch(`${env.API_URL}/media/${params.key}`, { method: request.method, headers });
	} catch {
		return new Response('Media unavailable', { status: 502 });
	}

	const responseHeaders = new Headers();
	for (const name of FORWARDED_RESPONSE_HEADERS) {
		const value = res.headers.get(name);
		if (value) responseHeaders.set(name, value);
	}
	return new Response(request.method === 'HEAD' ? null : res.body, {
		status: res.status,
		headers: responseHeaders
	});
};

export const GET = relay;
export const HEAD = relay;
