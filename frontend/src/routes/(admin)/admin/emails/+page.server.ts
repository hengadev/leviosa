import * as env from '$app/env/private';
import { isRedirect, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

// The email preview catalogue: every email the platform sends, rendered by the
// backend with sample data. Fetched here only; the browser never calls the
// backend for it.

export type EmailAudience = 'client' | 'partner' | 'practice';
export type EmailLocale = 'fr' | 'en';

export interface EmailEntry {
	id: string;
	label: string;
	audience: EmailAudience;
	locales: EmailLocale[];
	subjects: Partial<Record<EmailLocale, string>>;
}

export interface EmailPreview {
	subject: string;
	html: string;
	attachments: string[];
}

const LOAD_ERROR = "Impossible de charger les emails. Veuillez réessayer.";
const PREVIEW_ERROR = "Impossible d'afficher cet email. Veuillez réessayer.";

const MOCK_ENTRIES: EmailEntry[] = [
	{
		id: 'booking_confirmation',
		label: 'Confirmation de réservation',
		audience: 'client',
		locales: ['fr', 'en'],
		subjects: { fr: 'Votre réservation est confirmée', en: 'Your booking is confirmed' }
	},
	{
		id: 'partner_new_booking',
		label: 'Nouvelle réservation',
		audience: 'partner',
		locales: ['fr'],
		subjects: { fr: 'Nouvelle réservation' }
	}
];

function mockPreview(entry: EmailEntry, locale: EmailLocale): EmailPreview {
	const subject = entry.subjects[locale] ?? entry.label;
	return {
		subject,
		html: `<!doctype html><html><head><meta charset="utf-8"></head><body style="font-family:sans-serif;padding:24px"><h1>${subject}</h1><p>${locale === 'fr' ? 'Aperçu fictif (mode démo).' : 'Mock preview (demo mode).'}</p><p><a href="https://example.com">Lien</a></p></body></html>`,
		attachments: entry.id === 'booking_confirmation' ? ['reservation.ics'] : []
	};
}

// Links in the email open in a new tab instead of navigating the frame.
function withBlankTarget(html: string): string {
	const base = '<base target="_blank">';
	return /<head[^>]*>/i.test(html) ? html.replace(/<head[^>]*>/i, (head) => head + base) : base + html;
}

// The email to show: the one in the URL, else the first client email, else the
// first one. Its locale: the one in the URL when the email has it, else its
// first locale.
function select(entries: EmailEntry[], url: URL): { entry: EmailEntry; locale: EmailLocale } | null {
	const wanted = url.searchParams.get('email');
	const entry =
		entries.find((e) => e.id === wanted) ??
		entries.find((e) => e.audience === 'client') ??
		entries[0];
	if (!entry) return null;
	const lang = url.searchParams.get('lang');
	const locale = entry.locales.find((l) => l === lang) ?? entry.locales[0] ?? 'fr';
	return { entry, locale };
}

export const load: PageServerLoad = async ({ fetch, cookies, url }) => {
	if (env.USE_MOCK_DATA === 'true') {
		const selected = select(MOCK_ENTRIES, url)!;
		const preview = mockPreview(selected.entry, selected.locale);
		return {
			entries: MOCK_ENTRIES,
			selectedId: selected.entry.id,
			locale: selected.locale,
			preview: { ...preview, html: withBlankTarget(preview.html) },
			error: null as string | null
		};
	}

	const headers = { Authorization: `Bearer ${cookies.get('session')}` };
	const denied = (res: Response) => res.status === 401 || res.status === 403;

	let entries: EmailEntry[];
	try {
		const res = await fetch(`${env.API_URL}/admin/emails/previews`, { headers });
		if (denied(res)) throw redirect(302, '/auth');
		if (!res.ok) throw new Error(`Failed to fetch email previews: ${res.status} ${res.statusText}`);
		entries = ((await res.json()) as EmailEntry[] | null) ?? [];
	} catch (err) {
		if (isRedirect(err)) throw err;
		console.error('Error loading email previews:', err);
		return { entries: [], selectedId: null, locale: 'fr' as EmailLocale, preview: null, error: LOAD_ERROR };
	}

	const selected = select(entries, url);
	if (!selected) {
		return { entries, selectedId: null, locale: 'fr' as EmailLocale, preview: null, error: null };
	}

	let preview: EmailPreview | null = null;
	let error: string | null = null;
	try {
		const res = await fetch(
			`${env.API_URL}/admin/emails/previews/${encodeURIComponent(selected.entry.id)}?locale=${selected.locale}`,
			{ headers }
		);
		if (denied(res)) throw redirect(302, '/auth');
		if (!res.ok) throw new Error(`Failed to render email preview: ${res.status} ${res.statusText}`);
		const body = (await res.json()) as EmailPreview;
		preview = { subject: body.subject, html: withBlankTarget(body.html), attachments: body.attachments ?? [] };
	} catch (err) {
		if (isRedirect(err)) throw err;
		console.error('Error rendering email preview:', err);
		error = PREVIEW_ERROR;
	}

	return { entries, selectedId: selected.entry.id, locale: selected.locale, preview, error };
};
