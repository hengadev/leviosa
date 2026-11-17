package domain

import "testing"

func TestCollectedAndRefundedCents(t *testing.T) {
	tests := []struct {
		name      string
		booking   Booking
		collected int
		refunded  int
	}{
		{
			name:      "deposit paid: the deposit, not the full price",
			booking:   Booking{PaymentStatus: PaymentStatusPaid, TotalPriceCents: 8000, AmountPaidOnlineCents: intPtr(2000)},
			collected: 2000,
		},
		{
			name:      "paid before the amount was recorded: the full price",
			booking:   Booking{PaymentStatus: PaymentStatusPaid, TotalPriceCents: 8000},
			collected: 8000,
		},
		{
			name:    "not paid: nothing",
			booking: Booking{PaymentStatus: PaymentStatusPending, TotalPriceCents: 8000, AmountPaidOnlineCents: intPtr(2000)},
		},
		{
			name:    "payment failed: nothing",
			booking: Booking{PaymentStatus: PaymentStatusFailed, TotalPriceCents: 8000},
		},
		{
			name:      "cancelled with a partial refund: the recorded refund",
			booking:   Booking{PaymentStatus: PaymentStatusPartiallyRefunded, TotalPriceCents: 8000, AmountPaidOnlineCents: intPtr(2000), RefundedAmountCents: intPtr(1000)},
			collected: 2000,
			refunded:  1000,
		},
		{
			name:      "refunded without a recorded amount: everything collected",
			booking:   Booking{PaymentStatus: PaymentStatusRefunded, TotalPriceCents: 8000, AmountPaidOnlineCents: intPtr(2000)},
			collected: 2000,
			refunded:  2000,
		},
		{
			name:      "disputed: collected, not refunded",
			booking:   Booking{PaymentStatus: PaymentStatusDisputed, TotalPriceCents: 8000, AmountPaidOnlineCents: intPtr(2000)},
			collected: 2000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.booking.CollectedCents(); got != tt.collected {
				t.Errorf("CollectedCents() = %d, want %d", got, tt.collected)
			}
			if got := tt.booking.RefundedCents(); got != tt.refunded {
				t.Errorf("RefundedCents() = %d, want %d", got, tt.refunded)
			}
		})
	}
}
