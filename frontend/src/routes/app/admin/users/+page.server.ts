import type { PageServerLoad, Action, Actions } from './$types';
import { redirect } from '@sveltejs/kit';
import { API_URL } from '$lib/envVariables';

// TODO: complete the user type so that it is easy.
type User = {
	id: string;
	email: string;
	role: string;
	lastname: string;
	firstname: string;
	gender: string;
	birthdate: string;
	telephone: string;
	address: string;
	city: string;
	postalcard: string;
};

export const load: PageServerLoad = async ({ locals, cookies }) => {
	if (!locals.user) {
		throw redirect(301, '/');
	}
	if (locals.user.role !== 'admin') {
		throw redirect(301, '/');
	}
	const sessionId = cookies.get('sessionId');
	// TODO: send something for the user to be identified.use the session id ?
	const res = await fetch(`${API_URL}/api/v1/admin/users`, {
		method: 'GET',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ sessionId: sessionId })
	});
	const users: User[] = await res.json();
	return { users };
};

// get delete put
// TODO: Find all the actions I want the admin to be able to do.
const deleteUser: Action = async ({ request }) => {
	const formData = request.formData();
	const userId = formData.get('userid');
	const res = await fetch(`${API_URL}/api/v1/admin/users`, {
		method: 'DELETE',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ userId })
	});
	if (res.ok) {
		console.log('user deleted');
	}
};

export const actions: Actions = { deleteUser };
