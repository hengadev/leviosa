// A file to change variables depending on environment mode

let GOOGLE_CLIENT_ID: string;
let GOOGLE_CLIENT_SECRET: string;
let HOSTNAME: string;
let API_URL: string;

if (import.meta.env.PROD) {
	// Production: Use environment variables set by GitHub Actions
	GOOGLE_CLIENT_ID = process.env.GOOGLE_CLIENT_ID;
	GOOGLE_CLIENT_SECRET = process.env.GOOGLE_CLIENT_SECRET;
	HOSTNAME = process.env.HOSTNAME;
	API_URL = process.env.API_URL;
} else {
	// Development: Import from .env files
	const env = await import('$env/static/private');
	GOOGLE_CLIENT_ID = env.GOOGLE_CLIENT_ID;
	GOOGLE_CLIENT_SECRET = env.GOOGLE_CLIENT_SECRET;
	HOSTNAME = env.HOSTNAME;
	API_URL = env.API_URL;
}

export { GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, HOSTNAME, API_URL };
