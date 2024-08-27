export const cookieName = 'session_token';

export type CookieParsed = {
	Name: string;
	Value: string;
	Expires: Date;
	HttpOnly: boolean;
	Secure: boolean;
};
