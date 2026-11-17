package domain

import "time"

// Money collected (finance figures). Leviosa's revenue is what clients paid
// online, on the day they paid: the deposit when the product has one (the
// remainder is paid at the session, outside Leviosa), the full price
// otherwise. A booking is paid within its payment hold, minutes after it is
// created (ADR-0018), so its creation is its payment date: no separate
// payment timestamp is stored.

// CollectedCents is what the client paid online for the booking: the amount
// recorded when its payment was confirmed, or its total price for a booking
// paid before that amount was recorded. 0 when no payment went through.
func (b *Booking) CollectedCents() int {
	switch b.PaymentStatus {
	case PaymentStatusPaid, PaymentStatusRefunded, PaymentStatusPartiallyRefunded, PaymentStatusDisputed:
	default:
		return 0
	}
	if b.AmountPaidOnlineCents != nil {
		return *b.AmountPaidOnlineCents
	}
	return b.TotalPriceCents
}

// RefundedCents is what was given back to the client: the amount recorded
// by the cancellation, or everything collected when the payment is refunded
// without a recorded amount (an admin refund). 0 when nothing was refunded.
func (b *Booking) RefundedCents() int {
	switch b.PaymentStatus {
	case PaymentStatusRefunded, PaymentStatusPartiallyRefunded:
	default:
		return 0
	}
	if b.RefundedAmountCents != nil {
		return *b.RefundedAmountCents
	}
	if b.PaymentStatus == PaymentStatusRefunded {
		return b.CollectedCents()
	}
	return 0
}

// PaidAt is the day the booking's payment counts on: its creation (see above).
func (b *Booking) PaidAt() time.Time {
	return b.CreatedAt
}

// NetCollectedCents is what the booking brought in and kept: collected minus
// refunded. A disputed payment counts nothing until the dispute closes, as
// in the financial summary.
func (b *Booking) NetCollectedCents() int {
	if b.PaymentStatus == PaymentStatusDisputed {
		return 0
	}
	return b.CollectedCents() - b.RefundedCents()
}
