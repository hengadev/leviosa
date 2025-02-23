// make that thing better by changing the value depending on the environment mode
export const API_URL = 'http://localhost:3500';

// TODO: make something so that I can return some error right
export function getAPIURL(): string {
	const env = import.meta.env.VITE_ENVIRONMENT;
	if (env === 'PRODUCTION') {
		return 'http://frontend:3500';
	} else if (env === 'DEVELOPPMENT') {
		return 'http://localhost:3500';
	}
	return '';
}
