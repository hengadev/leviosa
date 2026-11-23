// The on-site payment (issue 039): the balance after a deposit, or the full
// price when nothing was paid online, is paid at the practice on a card
// terminal that is not connected to Leviosa. The practitioner records it in
// one tap ("Reste payé · N €") and can undo it the same Paris day. The
// backend is the source of truth for every amount and rule; these helpers
// only shape what Aujourd'hui and Réservations show.

/** The money fields of a booking, as the partner bookings endpoint returns them. */
export interface BookingMoney {
	status: string;
	payment_status: string;
	amount_paid_online_cents?: number;
	amount_due_on_site_cents?: number;
	amount_paid_on_site_cents?: number | null;
	paid_on_site_at?: string | null;
}

/** What the day's cards and the bookings list need about the on-site payment. */
export interface OnSitePayment {
	/** The payment state in French, e.g. "Acompte payé, reste 40 €". */
	label: string;
	/** What is due at the practice (price minus what was paid online). */
	dueCents: number;
	/** "40 €", for the "Reste payé · 40 €" button. */
	dueLabel: string;
	/** Whether the on-site payment is recorded. */
	recorded: boolean;
	/** Whether "Reste payé" may be recorded once the session has started:
	 * a confirmed or completed booking with money due, not yet recorded
	 * (never on a no-show: its deposit stays as it is). */
	recordable: boolean;
	/** Whether the recorded payment can still be undone: recorded today in Paris. */
	undoable: boolean;
}

const TIME_ZONE = 'Europe/Paris';

/** The calendar day (YYYY-MM-DD) of an instant in Paris. */
function parisDay(date: Date): string {
	return new Intl.DateTimeFormat('en-CA', {
		timeZone: TIME_ZONE,
		year: 'numeric',
		month: '2-digit',
		day: '2-digit'
	}).format(date);
}

/** "40 €", or "40,50 €" when there are cents. */
export function formatEuros(cents: number): string {
	const euros = cents / 100;
	return `${euros.toLocaleString('fr-FR', {
		minimumFractionDigits: Number.isInteger(euros) ? 0 : 2,
		maximumFractionDigits: 2
	})} €`;
}

/** The on-site payment state of a booking, at `now`. */
export function onSitePayment(booking: BookingMoney, now: Date = new Date()): OnSitePayment {
	const dueCents = Math.max(0, booking.amount_due_on_site_cents ?? 0);
	const paidOnlineCents = booking.amount_paid_online_cents ?? 0;
	const recorded = booking.amount_paid_on_site_cents != null;
	const paidOnSiteAt = booking.paid_on_site_at ? new Date(booking.paid_on_site_at) : null;

	return {
		label: paymentLabel(booking, dueCents, paidOnlineCents, recorded),
		dueCents,
		dueLabel: formatEuros(dueCents),
		recorded,
		recordable:
			(booking.status === 'confirmed' || booking.status === 'completed') && dueCents > 0 && !recorded,
		undoable: recorded && paidOnSiteAt !== null && parisDay(paidOnSiteAt) === parisDay(now)
	};
}

function paymentLabel(
	booking: BookingMoney,
	dueCents: number,
	paidOnlineCents: number,
	recorded: boolean
): string {
	switch (booking.payment_status) {
		case 'failed':
			return 'Paiement échoué';
		case 'refunded':
			return 'Remboursé';
		case 'partially_refunded':
			return 'Remboursé en partie';
		case 'disputed':
			return 'Paiement contesté';
	}
	if (recorded) return paidOnlineCents > 0 ? 'Reste payé' : 'Payé';
	if (dueCents === 0) return 'Payé';
	// A no-show keeps its deposit, and nothing more is asked.
	if (booking.status === 'no_show') return paidOnlineCents > 0 ? 'Acompte payé' : '';
	if (paidOnlineCents > 0) return `Acompte payé, reste ${formatEuros(dueCents)}`;
	return `À régler sur place · ${formatEuros(dueCents)}`;
}

/** The backend's refusals (409), as short French messages. */
const CONFLICT_MESSAGES: Record<string, string> = {
	'the on-site payment is already recorded': 'Le paiement est déjà enregistré.',
	'an on-site payment can only be recorded once the session has started':
		"La séance n'a pas encore commencé.",
	'nothing is due at the practice for this booking': 'Rien à régler sur place.',
	'an on-site payment can only be recorded on a confirmed or completed booking':
		'Impossible pour cette séance.',
	'no on-site payment is recorded for this booking': 'Aucun paiement à annuler.',
	'an on-site payment can only be undone on the same day it was recorded':
		"Trop tard : l'annulation n'est possible que le jour même."
};

/** Records (POST) or undoes (DELETE) the on-site payment of a booking.
 * Returns null on success, or a short French message to show. */
export async function changeOnSitePayment(
	bookingId: string,
	action: 'record' | 'undo'
): Promise<string | null> {
	try {
		const res = await fetch(`/api/bookings/${bookingId}/on-site-payment`, {
			method: action === 'record' ? 'POST' : 'DELETE'
		});
		if (res.ok) return null;
		if (res.status === 409) {
			const body = await res.json().catch(() => null);
			const message = typeof body?.error === 'string' ? CONFLICT_MESSAGES[body.error] : undefined;
			return message ?? 'Action impossible pour cette séance.';
		}
		if (res.status === 404) return 'Réservation introuvable.';
		return 'Une erreur est survenue. Veuillez réessayer.';
	} catch {
		return 'Erreur réseau. Veuillez réessayer.';
	}
}
