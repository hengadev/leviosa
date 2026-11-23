package domain

import "time"

// On-site payment. The balance after a deposit (or the full price when
// nothing was paid online) is paid at the practice, on a card terminal that
// is not connected to Leviosa. The practitioner records that it was paid:
// a confirmation only, the amount is computed, never typed.

// PaidOnlineCents is what the client paid online, as the amount due at the
// practice counts it: the recorded amount, or, for a booking paid online
// before that amount was recorded, its whole price (as CollectedCents counts
// it).
func (b *Booking) PaidOnlineCents() int {
	if b.AmountPaidOnlineCents != nil {
		return *b.AmountPaidOnlineCents
	}
	return b.CollectedCents()
}

// AmountDueOnSiteCents is what is due at the practice: the total price
// minus what was paid online; 0 when nothing is due. It does not change once
// the on-site payment is recorded (see AmountPaidOnSiteCents).
func (b *Booking) AmountDueOnSiteCents() int {
	due := b.TotalPriceCents - b.PaidOnlineCents()
	if due < 0 {
		return 0
	}
	return due
}

// RecordOnSitePayment records that the client paid the amount due at the
// practice. Allowed on a confirmed or completed booking whose session has
// started and that still has money due; never on a no-show (its deposit
// stays as it is), a cancelled, awaiting or expired booking, nor twice.
func (b *Booking) RecordOnSitePayment(now time.Time) error {
	if b.Status != BookingStatusConfirmed && b.Status != BookingStatusCompleted {
		return ErrOnSitePaymentNotAllowed
	}
	if b.AmountPaidOnSiteCents != nil {
		return ErrOnSitePaymentAlreadyRecorded
	}
	if now.Before(b.SlotStartTime) {
		return ErrOnSitePaymentBeforeStart
	}
	due := b.AmountDueOnSiteCents()
	if due <= 0 {
		return ErrNothingDueOnSite
	}

	b.AmountPaidOnSiteCents = &due
	b.PaidOnSiteAt = &now
	b.UpdatedAt = now
	return nil
}

// UndoOnSitePayment removes a recorded on-site payment, undoing a wrong tap.
// Allowed only on the same Europe/Paris day it was recorded: after that, the
// day's records are final.
func (b *Booking) UndoOnSitePayment(now time.Time) error {
	if b.AmountPaidOnSiteCents == nil {
		return ErrOnSitePaymentNotRecorded
	}
	if b.PaidOnSiteAt == nil ||
		civilDate(b.PaidOnSiteAt.In(PracticeLocation)) != civilDate(now.In(PracticeLocation)) {
		return ErrOnSitePaymentUndoSameDayOnly
	}

	b.AmountPaidOnSiteCents = nil
	b.PaidOnSiteAt = nil
	b.UpdatedAt = now
	return nil
}
