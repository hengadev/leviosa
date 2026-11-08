import { defineEnvVars } from '@sveltejs/kit/env';

// Every environment variable the app reads (SvelteKit 3 exposes only the
// variables declared here). All are dynamic: read from the container's
// environment when the server starts, never inlined at build time, so one
// image runs with either stack's env file. All are optional, and an unset
// variable stays undefined, as `$env/dynamic/*` behaved: the code applies its
// own fallbacks (e.g. CLIENT_IP_HEADER ?? 'x-client-ip').
const optional = (input: string | undefined) => input;

export const variables = defineEnvVars({
	API_URL: { schema: optional },
	COMING_SOON: { schema: optional },
	CLIENT_IP_HEADER: { schema: optional },
	COMMIT_SHA: { schema: optional },
	USE_MOCK_DATA: { schema: optional },
	MOCK_ADMIN_EMAIL: { schema: optional },
	MOCK_ADMIN_PASSWORD: { schema: optional },
	MOCK_PARTNER_EMAIL: { schema: optional },
	MOCK_PARTNER_PASSWORD: { schema: optional },
	MOCK_CLIENT_EMAIL: { schema: optional },
	MOCK_CLIENT_PASSWORD: { schema: optional },

	// Public: sent to the browser with the page.
	PUBLIC_STRIPE_PUBLISHABLE_KEY: { public: true, schema: optional },
	PUBLIC_PLAUSIBLE_DOMAIN: { public: true, schema: optional },
	PUBLIC_PLAUSIBLE_SRC: { public: true, schema: optional }
});
