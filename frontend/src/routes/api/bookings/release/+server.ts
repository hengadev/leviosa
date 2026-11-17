import * as env from '$app/env/private';
import type { RequestHandler } from './$types';

// POST /api/bookings/release?token=… — the booking card frees its slot when
// the client leaves the payment step (changes slot, closes the page, comes
// back from a failed redirect payment) instead of keeping it until the
// payment hold elapses (ADR-0018). The backend cancels the PaymentIntent and
// expires the booking; it keeps the hold (409) when a payment is under way.
export const POST: RequestHandler = async ({ url, fetch }) => {
    const token = url.searchParams.get('token');
    if (!token) return new Response(null, { status: 400 });

    try {
        const res = await fetch(`${env.API_URL}/bookings/release-hold?token=${encodeURIComponent(token)}`, {
            method: 'POST',
        });
        if (res.ok) return new Response(null, { status: 204 });
        return new Response(null, { status: res.status === 409 || res.status === 401 ? res.status : 502 });
    } catch {
        return new Response(null, { status: 503 });
    }
};
