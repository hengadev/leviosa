package domain

import (
	"testing"
	"time"
)

func TestComputeCancellationRefund(t *testing.T) {
	slot := time.Date(2026, 11, 10, 14, 0, 0, 0, time.UTC)
	tiers := []RefundTier{
		{MinHoursBefore: 48, RefundPercentage: 100},
		{MinHoursBefore: 24, RefundPercentage: 50},
		{MinHoursBefore: 0, RefundPercentage: 0},
	}
	before := func(d time.Duration) time.Time { return slot.Add(-d) }

	tests := []struct {
		name        string
		actor       CancellationActor
		tiers       []RefundTier
		paid        int
		cancelledAt time.Time
		wantAmount  int
		wantPercent int
		wantTier    *int // MinHoursBefore of the tier that applied
	}{
		{"well before the first tier", CancelledByClient, tiers, 3000, before(72 * time.Hour), 3000, 100, intPtr(48)},
		{"exactly at the 48 h boundary", CancelledByClient, tiers, 3000, before(48 * time.Hour), 3000, 100, intPtr(48)},
		{"just after the 48 h boundary", CancelledByClient, tiers, 3000, before(48*time.Hour + time.Second), 3000, 100, intPtr(48)},
		{"just before the 48 h boundary", CancelledByClient, tiers, 3000, before(48*time.Hour - time.Second), 1500, 50, intPtr(24)},
		{"exactly at the 24 h boundary", CancelledByClient, tiers, 3000, before(24 * time.Hour), 1500, 50, intPtr(24)},
		{"just before the 24 h boundary", CancelledByClient, tiers, 3000, before(24*time.Hour - time.Second), 0, 0, intPtr(0)},
		{"just after the 24 h boundary", CancelledByClient, tiers, 3000, before(24*time.Hour + time.Second), 1500, 50, intPtr(24)},
		{"exactly at the start", CancelledByClient, tiers, 3000, slot, 0, 0, intPtr(0)},
		{"after the start", CancelledByClient, tiers, 3000, slot.Add(time.Minute), 0, 0, nil},
		{"no tiers configured", CancelledByClient, nil, 3000, before(time.Hour), 3000, DefaultRefundPercentage, nil},
		{"practitioner cancellation, late", CancelledByPractitioner, tiers, 3000, before(time.Hour), 3000, 100, nil},
		{"practitioner cancellation, after the start", CancelledByPractitioner, tiers, 3000, slot.Add(time.Hour), 3000, 100, nil},
		{"nothing paid online", CancelledByClient, tiers, 0, before(72 * time.Hour), 0, 100, intPtr(48)},
		{"rounding half up", CancelledByClient, []RefundTier{{MinHoursBefore: 0, RefundPercentage: 50}}, 1001, before(time.Hour), 501, 50, intPtr(0)},
		{"rounding down", CancelledByClient, []RefundTier{{MinHoursBefore: 0, RefundPercentage: 33}}, 1001, before(time.Hour), 330, 33, intPtr(0)},
		{"rounding up", CancelledByClient, []RefundTier{{MinHoursBefore: 0, RefundPercentage: 67}}, 1001, before(time.Hour), 671, 67, intPtr(0)},
		{"tiers in any order", CancelledByClient, []RefundTier{tiers[2], tiers[0], tiers[1]}, 3000, before(30 * time.Hour), 1500, 50, intPtr(24)},
		{"tiers not reaching 0 h", CancelledByClient, tiers[:2], 3000, before(time.Hour), 0, 0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeCancellationRefund(tt.actor, tt.tiers, tt.paid, slot, tt.cancelledAt)
			if got.AmountCents != tt.wantAmount {
				t.Errorf("amount = %d, want %d", got.AmountCents, tt.wantAmount)
			}
			if got.Percentage != tt.wantPercent {
				t.Errorf("percentage = %d, want %d", got.Percentage, tt.wantPercent)
			}
			switch {
			case tt.wantTier == nil && got.Tier != nil:
				t.Errorf("tier = %+v, want none", *got.Tier)
			case tt.wantTier != nil && got.Tier == nil:
				t.Errorf("tier = none, want the %d h tier", *tt.wantTier)
			case tt.wantTier != nil && got.Tier.MinHoursBefore != *tt.wantTier:
				t.Errorf("tier = %d h, want %d h", got.Tier.MinHoursBefore, *tt.wantTier)
			}
		})
	}
}

func TestComputeCancellationRefund_NeverExceedsThePayment(t *testing.T) {
	slot := time.Date(2026, 11, 10, 14, 0, 0, 0, time.UTC)
	for paid := 0; paid <= 1000; paid += 7 {
		for pct := 0; pct <= 100; pct += 3 {
			got := ComputeCancellationRefund(CancelledByClient, []RefundTier{{MinHoursBefore: 0, RefundPercentage: pct}}, paid, slot, slot.Add(-time.Hour))
			if got.AmountCents < 0 || got.AmountCents > paid {
				t.Fatalf("paid %d at %d%%: refund %d out of [0, %d]", paid, pct, got.AmountCents, paid)
			}
		}
	}
}

func intPtr(v int) *int { return &v }
