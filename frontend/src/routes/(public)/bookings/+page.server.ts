import * as env from '$app/env/private';
import type { Actions, PageServerLoad } from './$types';

/**
 * The booking as the booking page shows it (GET /bookings/lookup, issue 017):
 * no internal IDs, no guest contact fields.
 */
export type PublicBooking = {
	id: string;
	product_name: string;
	partner_name: string;
	slot_start_time: string;
	slot_end_time: string;
	status: 'awaiting_payment' | 'expired' | 'confirmed' | 'cancelled' | 'completed' | 'no_show';
	total_price_cents: number;
	currency: string;
	payment_status: string;
	locale: 'fr' | 'en';
	amount_paid_online_cents: number;
	/** What a cancellation made now would refund; set while the booking can be cancelled. */
	refund_preview?: {
		amount_cents: number;
		percentage: number;
		/** Threshold of the cancellation tier that applies; absent without tiers. */
		min_hours_before?: number;
	};
	cancelled_by?: 'client' | 'practitioner';
	cancelled_at?: string;
	refunded_amount_cents?: number;
};

/** Why a cancellation failed, shown in the booking's language by the page. */
export type CancelFailure = 'missing' | 'not_cancellable' | 'failed';

export const load: PageServerLoad = async ({ url, fetch }) => {
	const token = url.searchParams.get('token');
	let booking: PublicBooking | null = null;
	let lookupError: string | null = null;

	// Token path: server-side lookup so the guest sees the booking immediately
	if (token) {
		try {
			const res = await fetch(`${env.API_URL}/bookings/lookup?token=${encodeURIComponent(token)}`);
			if (res.ok) {
				booking = await res.json();
			} else {
				lookupError = 'Ce lien est invalide ou a expiré.';
			}
		} catch {
			lookupError = 'Erreur réseau. Veuillez réessayer.';
		}
	}

	return { booking, lookupError, token };
};

export const actions: Actions = {
	lookup: async ({ request, fetch }) => {
		const formData = await request.formData();
		const ref = (formData.get('ref') as string)?.trim() ?? '';
		const email = (formData.get('email') as string)?.trim() ?? '';
		const phone = (formData.get('phone') as string)?.trim() ?? '';

		if (!ref) {
			return { action: 'lookup' as const, success: false, error: 'Veuillez entrer la référence de réservation.' };
		}

		const params = new URLSearchParams({ ref });
		if (email) {
			params.set('email', email);
		} else if (phone) {
			params.set('phone', phone);
		} else {
			return { action: 'lookup' as const, success: false, error: 'Veuillez entrer votre email ou numéro de téléphone.' };
		}

		try {
			const res = await fetch(`${env.API_URL}/bookings/lookup?${params}`);
			if (res.ok) {
				return { action: 'lookup' as const, success: true, booking: (await res.json()) as PublicBooking };
			}
			return { action: 'lookup' as const, success: false, error: 'Aucune réservation ne correspond à ces informations.' };
		} catch {
			return { action: 'lookup' as const, success: false, error: 'Erreur réseau. Veuillez réessayer.' };
		}
	},

	// The client cancellation sends no reason: no free-text field in the
	// cancellation flow (ADR-0019). The refund is the one the page previewed.
	cancel: async ({ request, fetch }) => {
		const formData = await request.formData();
		const bookingId = (formData.get('booking_id') as string)?.trim() ?? '';
		const token = (formData.get('token') as string)?.trim() ?? '';

		if (!bookingId || !token) {
			return { action: 'cancel' as const, success: false, failure: 'missing' as CancelFailure };
		}

		try {
			const res = await fetch(
				`${env.API_URL}/bookings/${encodeURIComponent(bookingId)}/cancel-public?token=${encodeURIComponent(token)}`,
				{ method: 'POST' }
			);
			if (res.ok) {
				return { action: 'cancel' as const, success: true, booking: (await res.json()) as PublicBooking };
			}
			const failure: CancelFailure = res.status === 400 ? 'not_cancellable' : 'failed';
			return { action: 'cancel' as const, success: false, failure };
		} catch {
			return { action: 'cancel' as const, success: false, failure: 'failed' as CancelFailure };
		}
	}
};
