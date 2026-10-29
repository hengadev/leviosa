package domain

import "time"

// CancellationActor is who cancelled a booking: the client (from the booking
// link or their account) or the practitioner (CONTEXT.md: Practitioner
// cancellation).
type CancellationActor string

const (
	CancelledByClient       CancellationActor = "client"
	CancelledByPractitioner CancellationActor = "practitioner"
)

// RefundTier is one step of a product's cancellation fee schedule (ADR-0014):
// a cancellation made at least MinHoursBefore hours before the appointment
// refunds RefundPercentage% of the amount paid online. It mirrors the
// catalog's CancellationTier, so the booking domain stays free of the
// catalog's types.
type RefundTier struct {
	MinHoursBefore   int
	RefundPercentage int
}

// CancellationRefund is what a cancellation refunds to the client.
type CancellationRefund struct {
	// AmountCents is refunded on the client's payment, rounded to the cent.
	AmountCents int
	// Percentage of the amount paid online that AmountCents represents.
	Percentage int
	// Tier is the tier that applied to a client cancellation; nil for a
	// practitioner cancellation, or when the product has no tiers.
	Tier *RefundTier
}

// DefaultRefundPercentage applies to a client cancellation when the product
// has no cancellation tiers configured. ADR-0014 does not define a default;
// without a published fee schedule nothing can be withheld, so the client
// gets the whole amount paid online back.
const DefaultRefundPercentage = 100

// ComputeCancellationRefund returns the refund of a booking cancelled at
// cancelledAt, for an appointment starting at slotStart, on which the client
// paid amountPaidOnlineCents online.
//
// A practitioner cancellation refunds everything. A client cancellation
// refunds the percentage of the tier matching the notice given (ADR-0014):
// tiers are ordered by descending MinHoursBefore and the first one whose
// threshold the notice reaches applies, a notice of exactly MinHoursBefore
// hours included (48 h before the slot falls in the ">= 48 h" tier). The
// notice is measured to the instant, not in whole hours. A notice below every
// threshold (a cancellation after the start, or a tier list that does not go
// down to 0 h) refunds nothing. With no tiers, DefaultRefundPercentage
// applies.
//
// The amount is rounded half up to the cent and never exceeds what was paid.
func ComputeCancellationRefund(
	actor CancellationActor,
	tiers []RefundTier,
	amountPaidOnlineCents int,
	slotStart, cancelledAt time.Time,
) CancellationRefund {
	if amountPaidOnlineCents < 0 {
		amountPaidOnlineCents = 0
	}

	if actor == CancelledByPractitioner {
		return refundOf(amountPaidOnlineCents, 100, nil)
	}
	if len(tiers) == 0 {
		return refundOf(amountPaidOnlineCents, DefaultRefundPercentage, nil)
	}

	// The tier with the highest threshold the notice reaches. Tiers are
	// validated in descending order, but the choice does not depend on it.
	notice := slotStart.Sub(cancelledAt)
	var matched *RefundTier
	for i := range tiers {
		if notice < time.Duration(tiers[i].MinHoursBefore)*time.Hour {
			continue
		}
		if matched == nil || tiers[i].MinHoursBefore > matched.MinHoursBefore {
			tier := tiers[i]
			matched = &tier
		}
	}
	if matched == nil {
		return CancellationRefund{AmountCents: 0, Percentage: 0}
	}
	return refundOf(amountPaidOnlineCents, matched.RefundPercentage, matched)
}

func refundOf(amountCents, percentage int, tier *RefundTier) CancellationRefund {
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}
	// Half up to the cent, in integers: amount × percentage / 100.
	refund := (amountCents*percentage + 50) / 100
	return CancellationRefund{AmountCents: refund, Percentage: percentage, Tier: tier}
}
