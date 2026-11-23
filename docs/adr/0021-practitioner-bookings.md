# ADR-0021: Practitioner Bookings — a Client Who Called

**Status**: Accepted
**Date**: 2026-11-23

---

## Context

The practitioner (an osteopath, who is also the administrator) takes calls during the practice day. A client who calls wants a session; the practitioner books it on the phone, between two sessions, and the client pays at the practice on the card terminal. Until now every booking came from the practice page: the client chose a slot, paid online (the deposit or the full price, ADR-0018), and the app emailed the confirmation, the 24-hour reminder and the practice access (the door code).

A booking made by the practitioner is different in three ways. The client agreed to it on the phone, so the app has nothing to tell them. Nothing is paid online. And the practitioner may book a slot the practice page would refuse: later today, or further ahead than clients may book.

## Decision

**A new use case beside the online `CreateBooking` (`CreatePractitionerBooking`), not a flag on it.**

- **Origin.** A booking records how it was made: `online` (the default, and every booking made before) or `practitioner`. It is set at creation and never changes (`bookings.origin`, `NOT NULL DEFAULT 'online'`; `Update` never writes it).
- **Slot rules: availability and overlap only.** The slot is checked by the slot engine as for an online booking (an open availability occurrence, exceptions excluded, on the grid, fitting, not overlapping another booking with its buffer time), but without the minimum notice and the booking horizon (`BookingPolicy.ForPractitioner`). A slot that has started is still refused. The form's free slots use the same switch (`GET /partners/booking-slots/{partnerId}`); the public slot endpoint is unchanged.
- **The same agenda lock.** The booking is written through the guarded insert under the partner's agenda lock (`CreateGuarded`), so a practitioner booking and an online booking of the same slot cannot both succeed.
- **Confirmed at once, paid at the practice.** No payment hold, no PaymentIntent, nothing paid online (`AmountPaidOnlineCents = 0`): the full price is due at the practice and shows as "À régler sur place · N €" until *Reste payé* is recorded (issue 039).
- **Nothing to the client.** No confirmation, no 24-hour reminder, no cancellation email, no practice-access update; no booking token is generated, so no booking link exists. The booking service never calls these messages for a practitioner booking, the reminder query leaves such bookings out, and the notification adapter refuses them as well (`BookingNotificationData.Origin`). The door code is never sent for one. The practitioner's own notifications follow ADR-0020: no "new booking" push for a booking they made, and the 15-minute session reminder still fires.
- **Cancelling it** goes through the existing practitioner cancellation: nothing was paid online, so nothing is refunded and Stripe is not called, and no one is emailed. The slot is offered again.
- **The client.** A guest entered by hand (first name, last name and phone required, email optional), encrypted with encx as every guest booking (ADR-0008), or an existing client (`client_id`, exclusive with the guest fields) for the search of issue 043. An email given is kept on the booking only.
- **Who may book.** The availability's partner, or an administrator. The products are the practitioner's: the published products of their profile (`product_ids`), or every published product when the profile lists none (the practice has one practitioner). The price is the product's active one-time price, as online.

## Alternatives considered

- **A flag on `CreateBooking`.** The online path holds a slot for a payment, creates a PaymentIntent and emits the confirmation; the practitioner path does none of it and drops two slot rules. A flag would thread "unless practitioner" through every step of the client-facing flow. Rejected.
- **A confirmation SMS or email to the client.** The client is on the phone with the practitioner; an extra message, and the door code in it, is noise the practitioner did not ask for. Rejected; the practitioner tells the client what they need.
- **Keeping the minimum notice for the practitioner.** The notice protects the practitioner from last-minute online bookings; it makes no sense against their own decision. Rejected.
- **Deriving the origin from the payment (no PaymentIntent and nothing paid online).** A product without a price due online would look the same. An explicit column says what happened. Rejected.

## Consequences

- Every client-facing message added later must skip `practitioner` bookings (check `Booking.IsPractitionerBooking` or `BookingNotificationData.Origin`).
- The partner bookings list carries the origin, so the agenda can say that cancelling a practitioner booking emails no one.
- Rolling back to a release before this one keeps the column (the migration is additive), but that release would send the 24-hour reminder to the clients of practitioner bookings with an email; roll back before those bookings' reminders are due, or accept it.
- A practitioner booking with an email can later be claimed by the client's account (booking claim), like any guest booking.
