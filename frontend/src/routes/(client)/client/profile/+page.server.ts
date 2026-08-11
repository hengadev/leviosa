import { env } from '$env/dynamic/private';
import { fail, redirect, type Actions } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, locals }) => {
	if (env.USE_MOCK_DATA === 'true') {
		return { user: locals.user ?? null };
	}

	const res = await fetch(`${env.API_URL}/users/me`);

	if (!res.ok) {
		return { user: null };
	}

	const user = await res.json();
	return { user };
};

export const actions: Actions = {
	// SvelteKit forbids mixing a `default` action with named actions on the
	// same page, so the pre-existing profile-update handler becomes a named
	// action alongside `deleteAccount` and `requestData` below.
	updateProfile: async ({ request, fetch }) => {
		const formData = await request.formData();

		const gender = formData.get('gender') as string;
		const birthdate = formData.get('birthdate') as string;
		const address1 = formData.get('address1') as string;
		const address2 = (formData.get('address2') as string) ?? '';
		const postalCode = formData.get('postalCode') as string;
		const city = formData.get('city') as string;

		const body: Record<string, unknown> = {};

		if (gender) body.gender = gender;
		if (birthdate) body.birthdate = birthdate + 'T00:00:00Z';
		if (address1) body.address1 = address1;
		if (address2) body.address2 = address2;
		if (postalCode) body.postal_code = postalCode;
		if (city) body.city = city;

		const res = await fetch(`${env.API_URL}/users/me`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body),
		});

		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			return fail(res.status, { error: err.message ?? 'Impossible de mettre à jour votre profil.' });
		}

		const updatedUser = await res.json();
		return { success: true as const, user: updatedUser };
	},

	// Issue 003: destructive, irreversible self-service account deletion,
	// wired to the existing authuser `DELETE /auth/me` endpoint.
	deleteAccount: async ({ fetch, cookies, locals }) => {
		if (env.USE_MOCK_DATA === 'true') {
			cookies.delete('mock_session', { path: '/' });
			throw redirect(302, '/auth');
		}

		let res: Response;
		try {
			res = await fetch(`${env.API_URL}/auth/me`, { method: 'DELETE' });
		} catch (error) {
			console.error('Delete account request failed:', error);
			return fail(503, {
				deleteError: 'Impossible de contacter le serveur. Veuillez réessayer.'
			});
		}

		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			return fail(res.status, {
				deleteError:
					err.error ?? err.message ?? 'Impossible de supprimer votre compte. Veuillez réessayer.'
			});
		}

		// The backend already revokes every session server-side; clear the
		// local cookies too so the browser is signed out immediately.
		cookies.delete(locals.sessionCookieName, { path: '/', domain: locals.cookieDomain });
		cookies.delete('leviosa_refresh_token', { path: '/', domain: locals.cookieDomain });

		throw redirect(302, '/auth');
	}
};
