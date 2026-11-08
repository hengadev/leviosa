import * as env from '$app/env/private';
import { redirect, error } from '@sveltejs/kit';
import { computePermissions } from '#lib/security/permissions.ts';

export const load = async ({ locals, parent, fetch }) => {
	const { user, permissions } = await parent();

	if (!user) {
		throw redirect(302, '/auth?redirect=' + encodeURIComponent('/admin'));
	}

	if (user.role !== 'administrator') {
		throw error(403, 'Accès non autorisé');
	}

	let unreadCount = 0;
	// Whether the admin is also the practitioner (a verified profile of their
	// own): the navigation then leads to their practitioner area.
	let selfPractitioner = false;
	if (env.USE_MOCK_DATA !== 'true') {
		const [unreadRes, ownProfileRes] = await Promise.all([
			fetch(`${env.API_URL}/threads/unread-count`).catch(() => null),
			fetch(`${env.API_URL}/partners/me`).catch(() => null)
		]);
		try {
			if (unreadRes?.ok) {
				const data = await unreadRes.json();
				unreadCount = data.unread_count ?? 0;
			}
		} catch {
			// unread count is non-critical
		}
		try {
			if (ownProfileRes?.ok) {
				const ownProfile = await ownProfileRes.json();
				selfPractitioner =
					ownProfile.stripe_account_status === 'active' && ownProfile.stripe_onboarding_complete === true;
			}
		} catch {
			// without it, the practitioner area link stays hidden
		}
	}

	return { user, permissions, unreadCount, selfPractitioner };
};
