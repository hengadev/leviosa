import { type RequestHandler } from '@sveltejs/kit';
import { googleOAuth } from '$lib/auth/lucia.server';
import { API_URL, INTERNAL_API_KEY } from '$env/static/private';
import { route } from '$lib/ROUTES';

import { parseCookie } from '$lib/scripts/parseCookie';
import { cookieName } from '$lib/types/cookie';

import {
	GOOGLE_OAUTH_STATE_COOKIE_NAME,
	GOOGLE_OAUTH_CODE_VERIFIER_COOKIE_NAME
} from '$lib/auth/utils.server';
import { OAuth2RequestError } from 'arctic';

type GoogleUser = {
	sub: string;
	name: string;
	given_name: string;
	family_name: string;
	picture: string;
	email: string;
	email_verified: boolean;
	locale: string;
};

export const GET: RequestHandler = async ({ url, cookies }) => {
	const code = url.searchParams.get('code');
	const state = url.searchParams.get('state');

	const storedState = cookies.get(GOOGLE_OAUTH_STATE_COOKIE_NAME);
	const storedCodeVerifier = cookies.get(GOOGLE_OAUTH_CODE_VERIFIER_COOKIE_NAME);

	if (!code || !state || !storedState || !storedCodeVerifier || state !== storedState) {
		return new Response('Invalid OAuth state or code verifier', {
			status: 400
		});
	}
	try {
		const tokens = await googleOAuth.validateAuthorizationCode(code, storedCodeVerifier);
		const googleUserResponse = await fetch('https://openidconnect.googleapis.com/v1/userinfo', {
			headers: {
				Authorization: `Bearer ${tokens.accessToken}`
			}
		});
		const googleUser = (await googleUserResponse.json()) as GoogleUser;
		if (!googleUser.email) {
			return new Response('No primary email address', {
				status: 400
			});
		}
		if (!googleUser.email_verified) {
			return new Response('Unverified email address', {
				status: 400
			});
		}

		const res = await fetch(
			'https://people.googleapis.com/v1/people/me?personFields=birthdays,phoneNumbers,addresses,genders',
			{
				method: 'GET',
				headers: {
					Authorization: `Bearer ${tokens.accessToken}`,
					'Content-Type': 'application/json'
				}
			}
		);
		if (!res.ok) {
			return new Response('fetch other google credentials', {
				status: 400
			});
		}
		const values = await res.json();

		console.log('I just sent to the golang backend the following:', { ...googleUser, ...values });
		//
		const dbUserResponse = await fetch(`${API_URL}/api/v1/oauth/google/user`, {
			method: 'POST',
			headers: {
				'X-API-KEY': INTERNAL_API_KEY,
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ ...googleUser, ...values })
		});
		if (dbUserResponse.status !== 201) {
			return new Response('user not found', {
				status: 302,
				headers: {
					Location: route('/')
				}
			});
		}

		const cookie = dbUserResponse.headers.get('Set-Cookie');
		if (!cookie) {
			throw new Error(`No cookie found with response status: ${dbUserResponse.statusText}`);
		}
		console.log('I get the following cookie brother:', cookie);
		const cookieParsed = parseCookie(cookie);
		cookies.set(cookieName, cookieParsed.Value, {
			path: '/',
			maxAge: 60 * 60 * 24 * 30,
			secure: import.meta.env.PROD,
			...cookieParsed
		});
		return new Response(null, {
			status: 302,
			headers: {
				Location: route('/app')
			}
		});
	} catch (err) {
		console.error('catching an error', err);
		if (err instanceof OAuth2RequestError) {
			return new Response(null, {
				status: 400
			});
		}
		return new Response(null, {
			status: 500
		});
	}
};
