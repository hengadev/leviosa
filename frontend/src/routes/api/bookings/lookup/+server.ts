import * as env from '$app/env/private';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// GET /api/bookings/lookup?token=… — the booking card polls this after the
// payment until the webhook has confirmed the booking (issue 014). It answers
// the statuses only; the booking page (/bookings) shows the details.
export const GET: RequestHandler = async ({ url, fetch }) => {
    const token = url.searchParams.get('token');
    if (!token) return json({ error: 'invalid' }, { status: 400 });

    try {
        const res = await fetch(`${env.API_URL}/bookings/lookup?token=${encodeURIComponent(token)}`);
        if (!res.ok) return json({ error: 'not_found' }, { status: res.status === 404 ? 404 : 502 });
        const booking = (await res.json()) as { status: string; payment_status: string };
        return json(
            { status: booking.status, paymentStatus: booking.payment_status },
            { headers: { 'Cache-Control': 'no-store' } },
        );
    } catch {
        return json({ error: 'unavailable' }, { status: 503 });
    }
};
