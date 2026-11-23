import type { PageServerLoad } from './$types';
import * as env from '$app/env/private';
import { error, redirect, isRedirect } from '@sveltejs/kit';

const TIME_ZONE = 'Europe/Paris';

/** A booking as the partner bookings endpoint returns it (the fields used here). */
interface BackendBooking {
	id: string;
	client_name: string;
	client_phone?: string;
	product_name: string;
	slot_start_time: string;
	slot_end_time: string;
	status: string;
	payment_status: string;
	total_price_cents: number;
	completed_at?: string;
	no_show_at?: string;
}

/** A transaction of the partner earnings endpoint: amount_cents is what was paid online. */
interface BackendEarningsTransaction {
	id: string;
	amount_cents: number;
}

/** The partner next-working-day endpoint's answer (the fields used here). */
interface BackendNextWorkingDay {
	date: string;
	bookings: BackendBooking[];
}

export interface TodaySession {
	id: string;
	startsAt: string;
	endsAt: string;
	startLabel: string;
	endLabel: string;
	clientName: string;
	/** Digits (and a leading +) only, ready for a tel: link; empty when unknown. */
	phone: string;
	/** What to show next to the phone button, as the client wrote it. */
	phoneLabel: string;
	productName: string;
	paymentLabel: string;
}

/** One of today's sessions already closed (completed or no-show). */
export interface TodayClosedSession {
	id: string;
	startLabel: string;
	endLabel: string;
	clientName: string;
	productName: string;
	/** "Terminée" or "Absent". */
	resultLabel: string;
	/** Whether the Annuler (reopen) link applies: the close happened today
	 * in Paris (the domain refuses a reopen after that day). */
	canUndo: boolean;
}

export type TodayData =
	| { state: 'not-set-up' }
	| {
			state: 'ready';
			dateLabel: string;
			/** Today's sessions that are not cancelled, past ones included. */
			sessionCount: number;
			/** Confirmed sessions of today, in order; the page drops those that have ended. */
			sessions: TodaySession[];
			/** Today's confirmed sessions that have ended, waiting to be closed
			 * (Terminée / Absent), in order. */
			toClose: TodaySession[];
			/** Today's sessions already closed ("Terminées aujourd'hui"), in order. */
			closed: TodayClosedSession[];
			/** The next working day and its sessions, when there is one; absent when
			 * the fetch failed, so the page still renders. */
			nextWorkingDay?: NextWorkingDay;
	  };

export interface NextWorkingDay {
	/** The day in French, e.g. "Samedi 28 novembre". */
	dayLabel: string;
	sessions: TodaySession[];
}

// ── Helpers (Europe/Paris, whatever the server's zone) ──────────

/** The calendar day (YYYY-MM-DD) of an instant in Paris. */
function parisDay(date: Date): string {
	return new Intl.DateTimeFormat('en-CA', {
		timeZone: TIME_ZONE,
		year: 'numeric',
		month: '2-digit',
		day: '2-digit'
	}).format(date);
}

function parisTime(date: Date): string {
	return new Intl.DateTimeFormat('fr-FR', {
		timeZone: TIME_ZONE,
		hour: '2-digit',
		minute: '2-digit'
	}).format(date);
}

function capitalize(text: string): string {
	return text.charAt(0).toUpperCase() + text.slice(1);
}

function dayLabel(date: Date): string {
	return capitalize(
		new Intl.DateTimeFormat('fr-FR', {
			timeZone: TIME_ZONE,
			weekday: 'long',
			day: 'numeric',
			month: 'long'
		}).format(date)
	);
}

/** "+33 6 12 34 56 78" becomes "+33612345678". */
function dialable(phone: string): string {
	const trimmed = phone.trim();
	const digits = trimmed.replace(/\D/g, '');
	if (digits.length === 0) return '';
	return trimmed.startsWith('+') ? `+${digits}` : digits;
}

/** A French number reads "06 12 34 56 78"; any other number is shown as given. */
function displayPhone(phone: string): string {
	const digits = dialable(phone);
	const national = digits.startsWith('+33') && digits.length === 12 ? `0${digits.slice(3)}` : digits;
	if (/^0\d{9}$/.test(national)) return national.replace(/(\d{2})(?=\d)/g, '$1 ');
	return phone;
}

function formatEuros(cents: number): string {
	const euros = cents / 100;
	return `${euros.toLocaleString('fr-FR', {
		minimumFractionDigits: Number.isInteger(euros) ? 0 : 2,
		maximumFractionDigits: 2
	})} €`;
}

/** The payment state shown on the day's cards. */
function paymentLabel(booking: BackendBooking, paidOnlineCents: number | undefined): string {
	switch (booking.payment_status) {
		case 'paid': {
			if (paidOnlineCents === undefined) return 'Paiement reçu';
			const left = booking.total_price_cents - paidOnlineCents;
			return left > 0 ? `Acompte payé, reste ${formatEuros(left)}` : 'Payé';
		}
		case 'pending':
			return 'Paiement en attente';
		case 'failed':
			return 'Paiement échoué';
		case 'refunded':
			return 'Remboursé';
		case 'partially_refunded':
			return 'Remboursé en partie';
		case 'disputed':
			return 'Paiement contesté';
		default:
			return '';
	}
}

/** A booking as a row of the day cards (today's sessions and the next
 * working day's alike). paidOnlineCents is what the client paid online. */
function toSession(b: BackendBooking, paidOnlineCents: number | undefined): TodaySession {
	const start = new Date(b.slot_start_time);
	const end = new Date(b.slot_end_time);
	const phoneLabel = (b.client_phone ?? '').trim();
	return {
		id: b.id,
		startsAt: start.toISOString(),
		endsAt: end.toISOString(),
		startLabel: parisTime(start),
		endLabel: parisTime(end),
		clientName: b.client_name || 'Client inconnu',
		phone: dialable(phoneLabel),
		phoneLabel: displayPhone(phoneLabel),
		productName: b.product_name || 'Produit inconnu',
		paymentLabel: paymentLabel(b, paidOnlineCents)
	};
}

// ── Load ─────────────────────────────────────────────────────────

export const load: PageServerLoad = async ({ locals, fetch }): Promise<TodayData> => {
	if (!locals.user?.id) {
		throw redirect(302, '/auth');
	}

	if (env.USE_MOCK_DATA === 'true') {
		return {
			state: 'ready',
			dateLabel: dayLabel(new Date()),
			sessionCount: 0,
			sessions: [],
			toClose: [],
			closed: []
		};
	}

	const partnerId = locals.user.id;

	try {
		// The administrator is the practitioner only with an active, verified
		// profile of their own (the rule of the admin layout's selfPractitioner).
		const profileRes = await fetch(`${env.API_URL}/partners/me`);
		if (profileRes.status === 401) {
			throw redirect(302, '/auth');
		}
		if (profileRes.status >= 500) {
			throw new Error(`Failed to fetch the partner profile: ${profileRes.status}`);
		}
		let isPractitioner = false;
		if (profileRes.ok) {
			const profile = await profileRes.json();
			isPractitioner =
				profile.stripe_account_status === 'active' && profile.stripe_onboarding_complete === true;
		}
		if (!isPractitioner) {
			return { state: 'not-set-up' };
		}

		const bookingsRes = await fetch(`${env.API_URL}/partners/bookings/${partnerId}`);
		if (bookingsRes.status === 401) {
			throw redirect(302, '/auth');
		}
		if (!bookingsRes.ok) {
			throw new Error(`Failed to fetch bookings: ${bookingsRes.status} ${bookingsRes.statusText}`);
		}
		const data = await bookingsRes.json();
		const bookings: BackendBooking[] = Array.isArray(data) ? data : [];

		const now = new Date();
		const today = parisDay(now);
		const isAppointment = (b: BackendBooking) =>
			b.status === 'confirmed' || b.status === 'completed' || b.status === 'no_show';

		const todays = bookings
			.filter((b) => isAppointment(b) && parisDay(new Date(b.slot_start_time)) === today)
			.sort(
				(a, b) => new Date(a.slot_start_time).getTime() - new Date(b.slot_start_time).getTime()
			);
		const confirmed = todays.filter((b) => b.status === 'confirmed');

		// What the client paid online (the deposit, or the whole price): only the
		// earnings endpoint says it, so ask for it only when there is a session.
		const paidOnline = new Map<string, number>();
		if (confirmed.length > 0) {
			try {
				const earningsRes = await fetch(`${env.API_URL}/partners/earnings/${partnerId}`);
				if (earningsRes.ok) {
					const earnings = await earningsRes.json();
					for (const t of (earnings?.transactions ?? []) as BackendEarningsTransaction[]) {
						paidOnline.set(t.id, t.amount_cents);
					}
				}
			} catch (err) {
				console.error('Error loading earnings:', err);
			}
		}

		const sessions: TodaySession[] = confirmed.map((b) => toSession(b, paidOnline.get(b.id)));
		const toClose: TodaySession[] = confirmed
			.filter((b) => new Date(b.slot_end_time).getTime() <= now.getTime())
			.map((b) => toSession(b, paidOnline.get(b.id)));

		// Today's sessions already closed: "Terminée" or "Absent", undoable
		// only when the close happened today in Paris (the reopen rule).
		const closed: TodayClosedSession[] = todays
			.filter((b) => b.status === 'completed' || b.status === 'no_show')
			.map((b) => {
				const session = toSession(b, undefined);
				const closedAt = b.status === 'completed' ? b.completed_at : b.no_show_at;
				return {
					id: session.id,
					startLabel: session.startLabel,
					endLabel: session.endLabel,
					clientName: session.clientName,
					productName: session.productName,
					resultLabel: b.status === 'completed' ? 'Terminée' : 'Absent',
					canUndo: closedAt !== undefined && parisDay(new Date(closedAt)) === today
				};
			});

		// The next working day (issue 037): its fetch is optional — a failure
		// still renders the day without the section.
		let nextWorkingDay: NextWorkingDay | undefined;
		try {
			const nextRes = await fetch(`${env.API_URL}/partners/next-working-day/${partnerId}`);
			if (nextRes.status === 401) {
				throw redirect(302, '/auth');
			}
			if (nextRes.ok) {
				const next = (await nextRes.json()) as BackendNextWorkingDay;
				if (next.date) {
					// Noon UTC of that date is that date in Paris, whatever the season.
					const day = new Date(`${next.date}T12:00:00Z`);
					const dayBookings = Array.isArray(next.bookings) ? next.bookings : [];
					nextWorkingDay = {
						dayLabel: dayLabel(day),
						sessions: dayBookings
							.sort(
								(a, b) =>
									new Date(a.slot_start_time).getTime() -
									new Date(b.slot_start_time).getTime()
							)
							.map((b) => toSession(b, paidOnline.get(b.id)))
					};
				}
			}
		} catch (err) {
			if (isRedirect(err)) throw err;
			console.error('Error loading the next working day:', err);
		}

		return {
			state: 'ready',
			dateLabel: dayLabel(now),
			sessionCount: todays.length,
			sessions,
			toClose,
			closed,
			nextWorkingDay
		};
	} catch (err) {
		if (isRedirect(err)) throw err;
		console.error('Error loading the day:', err);
		throw error(503, 'Impossible de charger vos séances. Veuillez réessayer.');
	}
};
