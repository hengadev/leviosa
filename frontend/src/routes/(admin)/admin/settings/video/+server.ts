import * as env from '$app/env/private';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// The practice page video upload (multipart, field "video", up to 50 MB).
// The body streams through to the backend as it arrives: the frontend
// container has 256 MB of memory and must not hold the file. This is the only
// route allowed a body over 512K (src/hooks.server.ts, BODY_SIZE_LIMIT in the
// Dockerfile). The backend checks the administrator role, the size and the
// format.
const MAX_UPLOAD_BYTES = 51 * 1024 * 1024;

export const POST: RequestHandler = async ({ request, fetch, locals }) => {
	if (locals.user?.role !== 'administrator') {
		return json({ error: 'forbidden' }, { status: 403 });
	}
	const contentType = request.headers.get('content-type') ?? '';
	if (!contentType.startsWith('multipart/form-data') || !request.body) {
		return json({ error: 'expected a multipart form' }, { status: 415 });
	}

	// The 50 MB file plus the form's overhead, as the backend allows. Larger
	// bodies are refused before any byte is relayed (adapter-node would cut
	// them off at BODY_SIZE_LIMIT mid-stream, which reads as a failed upload).
	const headers: Record<string, string> = { 'Content-Type': contentType };
	const length = request.headers.get('content-length');
	if (length && Number(length) > MAX_UPLOAD_BYTES) {
		return json({ error: 'video file exceeds 50 MB' }, { status: 413 });
	}
	if (length) headers['Content-Length'] = length;

	let res: Response;
	try {
		res = await fetch(`${env.API_URL}/admin/settings/practice-page/video/upload`, {
			method: 'POST',
			headers,
			body: request.body,
			// Node's fetch needs this to send a stream as the body.
			duplex: 'half'
		} as RequestInit);
	} catch {
		return json({ error: 'upload failed' }, { status: 502 });
	}

	const body = await res.json().catch(() => ({}));
	return json(body, { status: res.status });
};
