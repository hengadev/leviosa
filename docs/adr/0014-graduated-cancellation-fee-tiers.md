# ADR-0014: Graduated Cancellation Fee Tiers (Supersedes ADR-0006's Cancellation Window)

**Status**: Accepted
**Date**: 2026-08-11

---

## Context

ADR-0006 established a binary cancellation rule: a booking cannot be cancelled at all within `N` hours of the session start time, where `N` is the product's `cancellation_window` (`CancellationHours`). Inside that window, `CancelBooking` rejects the request outright with `ErrCancellationWindowClosed`.

The business owner has approved moving away from this binary block. It gives clients no way to cancel a booking they genuinely cannot attend once the window closes, and it gives the business no compensation for the lost slot — a graduated fee, proportional to how late the cancellation is, serves both sides better: clients can always cancel, and the business is compensated according to how much notice it got.

This ADR is written per issue `docs/issues/016-cancellation-policy-tiers-config.md` and its parent PRD `docs/prd/booking-lifecycle-policy-engine.md`, which explicitly calls out that this reverses documented ADR-0006 behavior and requires an accompanying ADR recording the supersession.

## Decision

**This ADR supersedes only the "Cancellation window" clause of ADR-0006.** All other ADR-0006 constraints — the booking status FSM, 10-minute slot alignment, buffer time, product duration bounds, room allocation types, and gap detection — remain unchanged and in force.

The product-level cancellation configuration is extended from a single `CancellationHours` threshold into an ordered list of **cancellation fee tiers**. Each tier maps a "hours before appointment" threshold to a refund percentage:

```
MinHoursBefore  RefundPercentage
>= 48h          100%
>= 24h          50%
<  24h          0%
```

Tiers are stored per product, ordered by descending `MinHoursBefore`, and must together cover the full range down to zero (i.e., the last tier's `MinHoursBefore` is `0`), so that a refund percentage can always be resolved for any cancellation time. `CancellationHours` is retained on the `Product` domain entity for backward compatibility (existing callers, migrations) but is superseded as the source of cancellation-blocking truth; tier configuration is now authoritative for computing the refund percentage.

Cancellation is **always allowed**, regardless of how close to the appointment it happens. The previously binary in-window block (`ErrCancellationWindowClosed`) is removed. What fee tier applies determines only how much is refunded, never whether cancellation itself is possible.

**Scope note**: this ADR and its implementation (issue 016) cover only the *configuration* of tiers — persistence, validation, and admin UI for defining a product's tier list. Actually computing the refund percentage from time-to-appointment and wiring it into `CancelBooking` to trigger a partial `RefundPayment` is issue 017 (`cancellation-policy-apply-and-refund`, Wave 2, currently blocked) and is explicitly out of scope here. `CancelBooking`'s enforcement logic is unchanged by this ADR alone; it will be updated when 017 lands.

## Consequences

**Positive:**
- Clients can always cancel a booking, at a cost that scales with notice given, instead of being locked out entirely.
- The business is compensated for late cancellations instead of absorbing the full loss or relying on manual refund decisions.
- Tier configuration follows the existing per-product configuration pattern, so admin tooling and validation are consistent with how other product fields are managed.

**Negative:**
- Product configuration now carries an additional ordered collection (tiers) that must be kept valid (non-overlapping, fully covering, percentages in `[0, 100]`) — more validation surface than a single integer.
- Until issue 017 ships, tier configuration is inert: `CancelBooking` does not yet consult it. This is an intentional two-issue split (016 configures, 017 applies), not an oversight.
- `CancellationHours` becomes a legacy field once 017 lands; it is not removed in this ADR to avoid a breaking migration ahead of 017's implementation.

**Future path**: Issue 017 will change `CancelBooking` to resolve the applicable tier from time-to-appointment, compute the refund amount, and call `RefundPayment` with that partial amount instead of a full refund — at which point `ErrCancellationWindowClosed`-style blocking is fully retired from the runtime path (it is already unreachable in intent per this ADR, but the code removal is 017's job).
