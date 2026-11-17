package booking

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
)

// financialSummaryMaxFetch is a safety cap on the number of rows fetched from
// the DB before in-memory aggregation. slot_start_time is encrypted at rest and
// cannot be filtered in SQL, so we must pull a bounded set.
const financialSummaryMaxFetch = 10000

func (s *BookingService) GetFinancialSummary(ctx context.Context, from, to time.Time) (*domain.FinancialSummaryResponse, error) {
	// Payments made in [from, to): the money collected, on the day it was
	// paid (domain.CollectedCents). A booking is paid minutes after it is
	// created, so created_at, a plain column, selects the range in SQL.
	// to is exclusive; the repository's bound is inclusive.
	lastInstant := to.Add(-time.Microsecond)
	repoFilter := ports.BookingFilter{
		PaymentStatus: []domain.PaymentStatus{
			domain.PaymentStatusPaid,
			domain.PaymentStatusRefunded,
			domain.PaymentStatusPartiallyRefunded,
			domain.PaymentStatusDisputed,
		},
		// Appointments only: a payment on a booking awaiting confirmation or
		// expired is not revenue yet (ADR-0018).
		ExcludeStatus:  domain.NonAppointmentStatuses(),
		CreatedAfter:   &from,
		CreatedBefore:  &lastInstant,
		OrderBy:        "created_at",
		OrderDirection: "desc",
		Limit:          financialSummaryMaxFetch,
	}

	allEncx, err := s.bookingRepo.List(ctx, repoFilter)
	if err != nil {
		return nil, fmt.Errorf("list bookings for financial summary: %w", err)
	}

	if len(allEncx) == financialSummaryMaxFetch {
		slog.WarnContext(ctx, "financial summary fetch hit safety cap; results may be incomplete",
			"cap", financialSummaryMaxFetch)
	}

	// Decrypt all bookings
	var decrypted []*domain.Booking
	for _, encx := range allEncx {
		b, err := domain.DecryptBookingEncx(ctx, s.crypto, encx)
		if err != nil {
			slog.WarnContext(ctx, "skipping booking that failed to decrypt",
				"booking_id", encx.ID,
				"err", err)
			continue
		}
		decrypted = append(decrypted, b)
	}

	// Compute aggregates. Gross is everything collected, refunds what was
	// given back, net the difference. Disputed payments are tracked separately
	// (DisputedCents) and kept out of gross and net until the dispute closes
	// and the payment status resolves back to paid or refunded — see the
	// FinancialSummary doc comment.
	var grossRevenueCents, refundsCents, disputedCents int
	for _, b := range decrypted {
		if b.PaymentStatus == domain.PaymentStatusDisputed {
			disputedCents += b.CollectedCents()
			continue
		}
		grossRevenueCents += b.CollectedCents()
		refundsCents += b.RefundedCents()
	}

	// Most recent payment first (the repository already orders by created_at).
	inRange := decrypted

	// Build enriched transaction rows.
	transactions := make([]domain.FinancialTransaction, 0, len(inRange))
	for _, b := range inRange {
		t := domain.FinancialTransaction{
			ID:            b.ID,
			SlotStartTime: b.SlotStartTime,
			PaidAt:        b.PaidAt(),
			AmountCents:   b.CollectedCents(),
			RefundedCents: b.RefundedCents(),
			PaymentStatus: b.PaymentStatus,
			BookingStatus: b.Status,
			ClientName:    "Utilisateur inconnu",
			PartnerName:   "Praticien inconnu",
			ProductName:   "Produit inconnu",
		}

		if s.authUserClient != nil {
			if b.IsGuestBooking() {
				t.ClientName = b.GuestDisplayName()
			} else {
				t.ClientName = s.resolveUserName(ctx, *b.ClientID, "Utilisateur inconnu")
			}
			t.PartnerName = s.resolveUserName(ctx, b.PartnerID, "Praticien inconnu")
		} else if b.IsGuestBooking() {
			t.ClientName = b.GuestDisplayName()
		}
		if s.productService != nil {
			t.ProductName = s.resolveProductName(ctx, b.ProductID)
		}

		transactions = append(transactions, t)
	}

	return &domain.FinancialSummaryResponse{
		Summary: domain.FinancialSummary{
			GrossRevenueCents: grossRevenueCents,
			RefundsCents:      refundsCents,
			DisputedCents:     disputedCents,
			NetRevenueCents:   grossRevenueCents - refundsCents,
		},
		Transactions: transactions,
	}, nil
}
