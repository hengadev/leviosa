import { GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, HOSTNAME } from '$lib/envVariables';
import { Google } from 'arctic';

const getRedirectURL = (provider: string): string =>
	`https://${HOSTNAME}/auth/oauth/${provider}/callback`;

// TODO: that part does not work in production, I need to use my domain name for that
const googleRedirectUrl = getRedirectURL('google');
export const googleOAuth = new Google(GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, googleRedirectUrl);
