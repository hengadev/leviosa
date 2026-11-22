import * as env from '$app/env/private';
import { redirect, error } from '@sveltejs/kit';
import { computePermissions } from '#lib/security/permissions.ts';

export const load = async ({ locals, fetch }) => {
	const user = locals.user;

	if (!user) {
		throw redirect(302, '/auth?redirect=' + encodeURIComponent('/staff'));
	}

	const permissions = computePermissions(user.role);

	if (!permissions.canAccessOps) {
		throw error(403, 'Accès non autorisé');
	}

	let unreadCount = 0;
	// Whether an administrator is also the practitioner (a verified profile of
	// their own): without it, the practitioner pages say the profile is not
	// set up. A partner reaching this area always has one.
	let selfPractitioner = user.role !== 'administrator';
	if (env.USE_MOCK_DATA !== 'true') {
		const [unreadRes, ownProfileRes] = await Promise.all([
			fetch(`${env.API_URL}/threads/unread-count`).catch(() => null),
			user.role === 'administrator' ? fetch(`${env.API_URL}/partners/me`).catch(() => null) : null
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
			// without it, the admin is taken for not the practitioner
		}
	}

	return { user, permissions, unreadCount, selfPractitioner };
};
