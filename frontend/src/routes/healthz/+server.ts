import { json } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import type { RequestHandler } from './$types';

// Commit SHA this image was built from (frontend/Dockerfile bakes it in as
// COMMIT_SHA), or "dev" for unbuilt runs (pnpm dev, pnpm preview) — never a
// fake version.
const frontendSHA = () => env.COMMIT_SHA || 'dev';

// The backend must answer (or refuse) quickly: this endpoint gates deploys and
// container healthchecks, so a hanging backend would hang them too.
const BACKEND_TIMEOUT_MS = 2000;

// Reports both apps' release SHAs so deploys and Promote can verify what is
// actually serving. 200 only when the backend answered and reported itself
// healthy; anything else (unreachable, timeout, non-2xx) is a non-200 so
// healthchecks and the deploy gate fail loudly.
export const GET: RequestHandler = async () => {
	try {
		const res = await fetch(`${env.API_URL}/health`, {
			signal: AbortSignal.timeout(BACKEND_TIMEOUT_MS)
		});
		if (!res.ok) {
			return json(
				{ frontend: frontendSHA(), backend: null, error: `backend /health returned ${res.status}` },
				{ status: 502 }
			);
		}

		const body = (await res.json()) as { status?: string; sha?: string };
		if (body.status !== 'ok') {
			return json(
				{ frontend: frontendSHA(), backend: body.sha ?? null, error: `backend reported ${body.status}` },
				{ status: 502 }
			);
		}

		return json({ frontend: frontendSHA(), backend: body.sha ?? null });
	} catch {
		return json(
			{ frontend: frontendSHA(), backend: null, error: 'backend unreachable' },
			{ status: 502 }
		);
	}
};
