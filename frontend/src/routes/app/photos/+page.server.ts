import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { API_URL } from '$env/static/private';

export const load: PageServerLoad = async ({ locals }) => {
	if (!locals.user) {
		throw redirect(301, '/');
	}
	if (locals.user.role !== 'admin') {
		throw redirect(301, '/');
	}
	const res = await fetch(`${API_URL}/api/v1/photos`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' }
		// body: JSON.stringify({ userId })
	});
	const photos = await res.json();
	return { photos };
};
