package domain

import "fmt"

// CancellationTier is one step of a product's graduated cancellation-fee
// schedule (issue 016 / ADR-0014, superseding ADR-0006's binary
// cancellation-window rule).
//
// A tier says: "if the booking is cancelled with at least MinHoursBefore
// hours' notice before the appointment, refund RefundPercentage% of the
// amount paid." A product's tiers are stored as an ordered list, sorted by
// descending MinHoursBefore; the applicable tier for a given
// time-to-appointment is the first one whose MinHoursBefore the remaining
// time satisfies.
//
// This type only carries and validates configuration. Resolving a tier from
// an actual cancellation time and triggering the corresponding partial
// refund is issue 017's responsibility (Wave 2, blocked) — see ADR-0014.
type CancellationTier struct {
	// MinHoursBefore is the minimum number of hours before the appointment
	// start time required for this tier to apply. Must be >= 0.
	MinHoursBefore int `json:"minHoursBefore"`
	// RefundPercentage is the percentage of the paid amount refunded when
	// this tier applies. Must be in [0, 100].
	RefundPercentage int `json:"refundPercentage"`
}

// ValidateCancellationTiers validates a product's ordered list of
// cancellation fee tiers.
//
// An empty/nil list is valid: it means the product has not opted into
// graduated cancellation fees yet, and today's binary CancellationHours
// field continues to be the only cancellation-related configuration for
// that product (until issue 017 changes CancelBooking's behavior).
//
// When tiers are configured, the list must:
//   - contain each RefundPercentage within [0, 100]
//   - contain each MinHoursBefore >= 0
//   - be strictly ordered by descending MinHoursBefore, with no duplicate
//     thresholds
//   - fully cover the range down to a cancellation made with no notice at
//     all — i.e., the last (lowest-threshold) tier must have
//     MinHoursBefore == 0 — so a refund percentage can always be resolved
//     for any cancellation time.
func ValidateCancellationTiers(tiers []CancellationTier) error {
	if len(tiers) == 0 {
		return nil
	}

	for i, tier := range tiers {
		if tier.RefundPercentage < 0 || tier.RefundPercentage > 100 {
			return fmt.Errorf("tier %d: refund percentage must be between 0 and 100, got %d", i, tier.RefundPercentage)
		}
		if tier.MinHoursBefore < 0 {
			return fmt.Errorf("tier %d: minimum hours before appointment cannot be negative, got %d", i, tier.MinHoursBefore)
		}
		if i > 0 && tier.MinHoursBefore >= tiers[i-1].MinHoursBefore {
			return fmt.Errorf("tier %d: tiers must be strictly ordered by descending minimum hours before appointment (tier %d has %dh, tier %d has %dh)", i, i-1, tiers[i-1].MinHoursBefore, i, tier.MinHoursBefore)
		}
	}

	last := tiers[len(tiers)-1]
	if last.MinHoursBefore != 0 {
		return fmt.Errorf("tiers must fully cover down to a cancellation with no notice: the last tier must have minimum hours before appointment of 0, got %d", last.MinHoursBefore)
	}

	return nil
}
