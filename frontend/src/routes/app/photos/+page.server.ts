import { redirect } from '@sveltejs/kit';
import type { PageLoadServer } from './$types';
import { API_URL } from '$env/static/private';

export const load: PageLoadServer = async ({ locals }) => {
	if (!locals.user) {
		throw redirect(301, '/');
	}
	if (locals.user.role !== 'admin') {
		throw redirect(301, '/');
	}
	const userId = locals.user.id;
	const res = await fetch(`${API_URL}/api/v1/photos`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ userId })
	});
	// TODO: I need some machine learning or something to just identify people in the database
	// Je peux faire une preidentification en regardant les metadonnes du fichier tout en sachant que chaque personne a un creneau horraire attribue (metadata)
	const photos = await res.json();
	return { photos };
};
