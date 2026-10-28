# ADR-0018: Bookings Await Payment Before Confirmation (Extends ADR-0006's Status FSM)

**Status**: Accepted
**Date**: 2026-10-27

---

## Context

ADR-0006's booking FSM starts at `confirmed`. `CreateBooking` persists the booking as `confirmed` and creates a Stripe PaymentIntent, but nothing on the public booking page ever collects the payment. A client who leaves at the payment step therefore leaves a confirmed, unpaid booking holding the slot forever.

The single-practitioner launch (leviosa.care) requires online payment at booking time, so the slot must be held while the client pays, and released if they don't.

## Decision

A booking starts in a new status, `awaiting_payment`, before `confirmed`:

```
awaiting_payment → confirmed
awaiting_payment → expired
confirmed → completed | cancelled | no_show   (unchanged)
```

- An `awaiting_payment` booking holds its slot for a **payment hold** of 15 minutes (configurable). It counts as taken when slots are derived.
- The Stripe webhook confirming the payment moves it to `confirmed`. Confirmation email/SMS are sent only then.
- When the hold elapses unpaid, the booking becomes `expired` and the slot is offered again.
- A payment that succeeds after expiry confirms the booking if the slot is still free; otherwise it is refunded automatically and the client is notified.
- Payment is collected with the Stripe Payment Element embedded in the booking page, on the PaymentIntent `CreateBooking` already creates, rather than a redirect to Stripe Checkout.
- The amount collected online is the product's deposit when one is set (the remainder is not collected by Leviosa), otherwise the full price.

## Alternatives considered

- **Create the booking only after payment succeeds.** No hold, so two clients can pay for the same slot; one must then be refunded. Rejected.
- **Stripe Checkout (redirect).** Less frontend work, but leaves the branded page and requires a second flow alongside the existing PaymentIntent. Rejected.

## Consequences

- Every query that treats bookings as occupying time must include `awaiting_payment`; every query about real appointments (agenda, earnings, reminders, reviews) must exclude it and `expired`.
- An expiry job (or expiry evaluated at read time) is required.
- `expired` bookings are never shown to the partner as appointments.
