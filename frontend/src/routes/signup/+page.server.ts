import { redirect } from '@sveltejs/kit';
import type { Action, PageServerLoad } from './$types';
import { API_URL } from '$env/static/private';

import { parseCookie } from '$lib/scripts/parseCookie';
import { validate } from '$lib/scripts/credentials';

export const actions: Action = {
	default: async ({ request, cookies }) => {
		const formData = await request.formData();
		const email = String(formData.get('email'));
		const password = String(formData.get('password'));
		validate(email, password);

		const firstname = String(formData.get('firstname'));
		const lastname = String(formData.get('lastname'));
		const telephone = String(formData.get('telephone'));

		// TODO: check all the different informations if they have a valid format.
		const body = JSON.stringify({ email, password, firstname, lastname, telephone });

		const res = await fetch(`${API_URL}/api/v1/signup`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body
		});
		if (res.ok) {
			const cookieParsed = parseCookie(res.headers.getSetCookie()[0]);
			cookies.set('sessionId', cookieParsed.sessionId, {
				path: '/'
			});
			// redirect to the auth home page.
			throw redirect(303, '/app');
		}

		return { success: false, message: 'Something went wrong during the sign up.' };
	}
};

export const load: PageServerLoad = ({ locals }) => {
	if (locals.user.role === 'admin') {
		throw redirect(301, '/app/admin');
	} else if (locals.user) {
		throw redirect(301, '/app');
	}
};
