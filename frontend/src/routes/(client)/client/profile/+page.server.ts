import { env } from '$env/dynamic/private';
import { fail, redirect, type Actions } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

interface ThreadSummary {
	thread_id: string;
	participant_id: string;
	participant_name: string;
	last_message: string;
	last_message_at: string;
	unread_count: number;
}

// Canned request body sent into the client↔partner thread when a client asks
// for their personal data. No new backend/inbound channel — this rides the
// existing messaging system per docs/prd/legal-compliance-gdpr-foundations.md.
const DATA_REQUEST_MESSAGE =
	"Bonjour,\n\nJe souhaite exercer mon droit d'accès à mes données personnelles (RGPD, article 15) et en recevoir une copie. Merci de me confirmer la réception de cette demande.\n\n— Demande envoyée automatiquement depuis mon espace client.";

const FALLBACK_COMPANY_EMAIL = 'contact@leviosa.care';

export const load: PageServerLoad = async ({ fetch, locals }) => {
	if (env.USE_MOCK_DATA === 'true') {
		return { user: locals.user ?? null, companyEmail: FALLBACK_COMPANY_EMAIL };
	}

	// Best-effort: this is a public settings endpoint (see
	// docs/prd/legal-compliance-gdpr-foundations.md). Used as the "no active
	// conversation" fallback contact for the RGPD data-request below — if it
	// fails, the hardcoded fallback keeps that path from breaking.
	const companyEmail = await fetch(`${env.API_URL}/settings/email`)
		.then((res) => (res.ok ? res.json() : null))
		.then((body) => body?.email ?? FALLBACK_COMPANY_EMAIL)
		.catch(() => FALLBACK_COMPANY_EMAIL);

	const res = await fetch(`${env.API_URL}/users/me`);

	if (!res.ok) {
		return { user: null, companyEmail };
	}

	const user = await res.json();
	return { user, companyEmail };
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
	},

	// Issue 004: "request my data" — routes into the existing client↔partner
	// messaging thread rather than building new export tooling. See
	// docs/runbooks/gdpr-data-export.md for how staff fulfill the request.
	requestData: async ({ fetch }) => {
		if (env.USE_MOCK_DATA === 'true') {
			return { dataRequestSent: true as const, threadId: 'c1' };
		}

		let threadsRes: Response;
		try {
			threadsRes = await fetch(`${env.API_URL}/threads`);
		} catch (error) {
			console.error('Fetch threads for data request failed:', error);
			return fail(503, {
				dataRequestError: 'Impossible de contacter le serveur. Veuillez réessayer.'
			});
		}

		if (!threadsRes.ok) {
			return fail(threadsRes.status, {
				dataRequestError: 'Impossible de récupérer vos conversations. Veuillez réessayer.'
			});
		}

		const threads: ThreadSummary[] = await threadsRes.json();
		if (!threads.length) {
			// No existing partner conversation to carry the request — surface
			// the manual email fallback instead of failing silently.
			return { dataRequestNoThread: true as const };
		}

		// Threads are returned most-recently-active first.
		const threadId = threads[0].thread_id;

		const sendRes = await fetch(`${env.API_URL}/threads/${threadId}/messages`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ body: DATA_REQUEST_MESSAGE })
		});

		if (!sendRes.ok) {
			const err = await sendRes.json().catch(() => ({}));
			return fail(sendRes.status, {
				dataRequestError:
					err.error ?? err.message ?? "Impossible d'envoyer votre demande. Veuillez réessayer."
			});
		}

		return { dataRequestSent: true as const, threadId };
	}
};
