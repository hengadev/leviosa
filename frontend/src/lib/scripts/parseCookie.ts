import { cookieName, type CookieParsed } from '$lib/types/cookie';

/**
 * Parse cookie into a CookieParsed object
 *
 * @param   cookie  the cookie in string format.
 * @returns A cookie parsed into an exploitable CookieParsed object.
 */
function parseCookie(cookie: string): CookieParsed {
	// the default cookie object
	const res: CookieParsed = {
		Name: cookieName,
		Value: '',
		Expires: new Date(),
		HttpOnly: true,
		Secure: true
	};
	cookie.split(';').map((field) => {
		const split = field.trim().split('=');
		if (split[0] === 'Expires') {
			res.Expires = new Date(split[1]);
		}
		if (split[0] === cookieName) {
			res.Value = split[1];
		}
	});
	return res;
}

export { parseCookie };
