import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import { FRENCH_PHONE, normalizePhone } from '$lib/phone';
import type { RequestHandler } from './$types';

// POST /api/bookings — the practice page's guest booking (issue 014,
// ADR-0018). The browser cannot reach the backend, so this forwards the
// booking card's details to POST /bookings and answers only what the card
// needs to collect the payment: the booking token (to poll the
// confirmation), the PaymentIntent's client secret, the amount charged
// online and the end of the payment hold. No free-text field is forwarded
// (ADR-0019).

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const EMAIL = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
type BackendCreated = {
    id: string;
    status: string;
    token?: string;
    client_secret?: string;
    amount_due_online_cents?: number;
    hold_expires_at?: string;
};

const text = (value: unknown, max: number) => (typeof value === 'string' ? value.trim().slice(0, max) : '');

export const POST: RequestHandler = async ({ request, fetch }) => {
    let body: Record<string, unknown>;
    try {
        body = await request.json();
    } catch {
        return json({ error: 'invalid' }, { status: 400 });
    }

    const availabilityId = text(body.availabilityId, 36);
    const productId = text(body.productId, 36);
    const slotStart = text(body.slotStart, 40);
    const firstName = text(body.firstName, 100);
    const lastName = text(body.lastName, 100);
    const email = text(body.email, 254);
    const phone = normalizePhone(text(body.phone, 30));
    const locale = body.locale === 'en' ? 'en' : 'fr';

    if (
        !UUID.test(availabilityId) ||
        !UUID.test(productId) ||
        Number.isNaN(Date.parse(slotStart)) ||
        !firstName ||
        !lastName ||
        !EMAIL.test(email) ||
        !FRENCH_PHONE.test(phone) ||
        body.acceptTerms !== true
    ) {
        return json({ error: 'invalid' }, { status: 400 });
    }

    let res: Response;
    try {
        res = await fetch(`${env.API_URL}/bookings`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                availability_id: availabilityId,
                product_id: productId,
                slot_start_time: slotStart,
                guest_first_name: firstName,
                guest_last_name: lastName,
                guest_email: email,
                guest_phone: phone,
                locale,
            }),
        });
    } catch {
        return json({ error: 'unavailable' }, { status: 503 });
    }

    if (res.status === 409) return json({ error: 'slot_taken' }, { status: 409 });
    if (res.status === 400) return json({ error: 'invalid' }, { status: 400 });
    if (!res.ok) return json({ error: 'unavailable' }, { status: 502 });

    const created = (await res.json()) as BackendCreated;
    return json(
        {
            token: created.token ?? '',
            status: created.status,
            clientSecret: created.client_secret ?? '',
            amountDueCents: created.amount_due_online_cents ?? 0,
            holdExpiresAt: created.hold_expires_at ?? null,
        },
        { status: 201 },
    );
};
