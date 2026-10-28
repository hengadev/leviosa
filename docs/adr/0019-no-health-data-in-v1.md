# ADR-0019: Leviosa Stores No Health Data

**Status**: Accepted
**Date**: 2026-10-27

---

## Context

The first practice on Leviosa (leviosa.care) is an osteopath. Free-text fields on a booking — `client_notes` (reason for the visit) and `partner_notes` (the partner agenda lets the partner edit them) — would naturally receive health information: symptoms, history, treatment given.

Health data is a special category under GDPR (art. 9), and in France hosting it for a third party requires an HDS-certified host. The single Hetzner VPS (ADR-0016) is not HDS-certified.

## Decision

Leviosa stores booking data only: who, when, which product, contact details, payment. It stores no reason for the visit, no symptoms and no clinical notes.

- `client_notes` and `partner_notes` are neither collected nor shown, and the API does not accept writes to them. The columns stay in the schema.
- No free-text field is offered to clients during booking or cancellation.
- The partner keeps clinical notes in their own tools, outside Leviosa. The privacy policy says so.

## Alternatives considered

- **Move to an HDS-certified host.** Several times the cost and a different deployment model (ADR-0016) for a feature the practice does not need from Leviosa. Rejected for now.
- **Keep the fields and tell the partner not to write health data.** Relies on discipline; one note makes the platform non-compliant. Rejected.

## Consequences

- Re-enabling any free-text field on a booking for a health practice requires revisiting this ADR and the hosting.
