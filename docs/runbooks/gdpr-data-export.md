# Runbook: fulfilling a client GDPR data-access/export request

**Audience:** staff/admin operators
**Related:** `docs/prd/legal-compliance-gdpr-foundations.md`, issue `004-client-data-export-request.md`

## When this applies

A client used the "Demander mes données" (Request my data) button on their
profile page (`(client)/profile`). This sends a message into the existing
client↔partner conversation thread (visible in the staff messaging inbox,
same as any other client message) asking to exercise their GDPR Article 15
right of access.

If a client has no existing thread with a partner yet, the UI shows them a
fallback instructing them to email `contact@leviosa.care` directly instead —
that request arrives as a normal support email, not through this flow.

## Legal deadline

**One calendar month** from the date the request is received (GDPR Article
12(3)). This can be extended by two further months for complex or numerous
requests, but the client must be informed of the extension and the reason
within the first month. Log the receipt date as soon as you see the message
so the deadline is trackable.

## Fulfillment steps

1. **Identify the client.** Open the conversation in the staff messaging
   inbox that contains the request; note the client's user ID / email from
   the thread participant.
2. **Pull the client's data via the admin `users` view** (`/admin/users`):
   - Profile fields: name, email, phone, address, gender, birthdate.
   - Booking history: past and upcoming reservations, notes, financial
     status (visible via the client's booking records / admin financial
     views).
   - Payment records: Stripe customer ID and invoice/receipt history
     (via the admin Stripe-linked views — do not include full card
     numbers, Stripe never exposes them).
   - Messaging history: the client's message threads (export the thread(s)
     they're a participant in; do not include other clients' unrelated
     threads).
3. **Compile the export.** A plain document (PDF or text) listing the data
   above, organized by category, is sufficient — there is no required
   machine format for a company this size (GDPR Article 20 doesn't mandate
   automated portability here; see the parent PRD for the reasoning).
4. **Send it back through the same channel** — reply in the same message
   thread with the export attached (or a secure link if the file is large),
   so there's a paper trail of both the request and the fulfillment inside
   the existing messaging system. Do not send personal data over a channel
   the client didn't request it through.
5. **Log the response date.** Note in the thread (or wherever staff track
   this — currently no dedicated audit log for GDPR requests) that the
   request was fulfilled and on what date, so the one-month SLA is
   demonstrably met if ever questioned by CNIL.

## If the deadline is at risk

Reply in-thread before the one-month mark explaining the delay and the
extended timeline (up to two more months), per Article 12(3). Do this even
if the export isn't ready — an on-time acknowledgment with a revised ETA
satisfies the obligation; silence past the deadline does not.

## Out of scope for this runbook

- Automated/self-service export tooling — deliberately not built (see PRD).
- Right-to-erasure requests: those are self-service already via the
  "Delete account" button on the client's profile page and need no manual
  staff action.
