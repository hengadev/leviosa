import { GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET } from '$env/static/private';
import { Google } from 'arctic';

const googleRedirectUrl = `https://localhost:5173/auth/oauth/google/callback`;
export const googleOAuth = new Google(GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, googleRedirectUrl);
