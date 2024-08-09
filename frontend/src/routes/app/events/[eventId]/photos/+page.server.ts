import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { API_URL } from '$env/static/private';

async function getPhotos(userId: string, eventId: string) {
	// TODO: fetch the photos from S3 bucket via the backend golang
	const body = JSON.stringify({ userId, eventId });
	const res = await fetch(`${API_URL}/api/v1/photos`, {
		method: 'GET',
		headers: { 'Content-Type': 'application/json' },
		body
	});
	const photos = res.json();
	// const photos = ["photo1", "photo2", "photo3"]
	return photos;
}

export const load: PageServerLoad = async ({ locals, params }) => {
	if (!locals.user) {
		throw redirect(301, '/');
	}
	// how to get the params of the page
	const userid = locals.user.id;
	const eventid = params.eventId;
	const photos = await getPhotos(userid, eventid);
	return { photos };
};
