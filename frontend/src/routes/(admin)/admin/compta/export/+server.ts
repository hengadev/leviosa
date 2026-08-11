import { env } from '$env/dynamic/private';
import { error, isRedirect, redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// GET /admin/compta/export?from=YYYY-MM-DD&to=YYYY-MM-DD
// Streams the CSV export of the financial-summary endpoint straight through
// to the browser as a file download. Lives under /admin so hooks.server.ts's
// auth middleware (which only guards the /admin, /staff, /client prefixes)
// enriches the request with the session cookie before it reaches the backend.
export const GET: RequestHandler = async ({ fetch, url }) => {
	const from = url.searchParams.get('from') ?? '';
	const to = url.searchParams.get('to') ?? '';

	const params = new URLSearchParams({ format: 'csv' });
	if (from) params.set('from', from);
	if (to) params.set('to', to);

	try {
		const res = await fetch(`${env.API_URL}/admin/bookings/financial-summary?${params.toString()}`);

		if (res.status === 401 || res.status === 403) {
			throw redirect(302, '/auth');
		}
		if (!res.ok) {
			throw new Error(`Failed to export financial summary: ${res.status} ${res.statusText}`);
		}

		const body = await res.arrayBuffer();
		return new Response(body, {
			status: 200,
			headers: {
				'Content-Type': res.headers.get('Content-Type') ?? 'text/csv; charset=utf-8',
				'Content-Disposition':
					res.headers.get('Content-Disposition') ?? 'attachment; filename="financial-summary.csv"'
			}
		});
	} catch (err) {
		if (isRedirect(err)) throw err;
		console.error('Error exporting financial summary CSV:', err);
		throw error(503, 'Impossible de générer l’export CSV. Veuillez réessayer.');
	}
};
