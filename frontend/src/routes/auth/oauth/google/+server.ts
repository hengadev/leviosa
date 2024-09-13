import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

import { generateCodeVerifier, generateState } from 'arctic';

import {
	GOOGLE_OAUTH_STATE_COOKIE_NAME,
	GOOGLE_OAUTH_CODE_VERIFIER_COOKIE_NAME
} from '$lib/auth/utils.server';

import { googleOAuth } from '$lib/auth/lucia.server';

export const GET: RequestHandler = async ({ cookies }) => {
	const state = generateState();
	const codeVerifier = generateCodeVerifier();

	const url = await googleOAuth.createAuthorizationURL(state, codeVerifier, {
		scopes: ['profile', 'email']
	});

	cookies.set(GOOGLE_OAUTH_STATE_COOKIE_NAME, state, {
		path: '/',
		secure: import.meta.env.PROD,
		httpOnly: true,
		maxAge: 60 * 10,
		sameSite: 'lax'
	});
	cookies.set(GOOGLE_OAUTH_CODE_VERIFIER_COOKIE_NAME, codeVerifier, {
		path: '/',
		secure: import.meta.env.PROD,
		httpOnly: true,
		maxAge: 60 * 10,
		sameSite: 'lax'
	});
	redirect(302, url);
};
