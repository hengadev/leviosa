package booking

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/google/uuid"
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

	// On-site payments count on the day they were recorded (Europe/Paris),
	// which can be a later month than the booking's creation (a deposit paid
	// online in October, its balance at the practice in November). The
	// created_at query above misses those, so a second query selects bookings
	// whose paid_on_site_at falls in the window (widened by a day on each side
	// to cover the Paris offset; the precise day is checked in Go). Undoing an
	// on-site payment clears paid_on_site_at, so the row disappears here
	// without any extra work.
	onSiteFilter := repoFilter
	onSiteFilter.CreatedAfter = nil
	onSiteFilter.CreatedBefore = nil
	onSiteFilter.PaidOnSiteAfter = &[]time.Time{from.Add(-24 * time.Hour)}[0]
	onSiteFilter.PaidOnSiteBefore = &[]time.Time{to.Add(24 * time.Hour)}[0]
	onSiteEncx, err := s.bookingRepo.List(ctx, onSiteFilter)
	if err != nil {
		return nil, fmt.Errorf("list on-site-paid bookings for financial summary: %w", err)
	}
	if len(onSiteEncx) == financialSummaryMaxFetch {
		slog.WarnContext(ctx, "financial summary on-site fetch hit safety cap; results may be incomplete",
			"cap", financialSummaryMaxFetch)
	}

	known := make(map[uuid.UUID]struct{}, len(decrypted))
	for _, b := range decrypted {
		known[b.ID] = struct{}{}
	}
	for _, encx := range onSiteEncx {
		if _, ok := known[encx.ID]; ok {
			continue
		}
		b, err := domain.DecryptBookingEncx(ctx, s.crypto, encx)
		if err != nil {
			slog.WarnContext(ctx, "skipping booking that failed to decrypt",
				"booking_id", encx.ID,
				"err", err)
			continue
		}
		decrypted = append(decrypted, b)
	}

	// Compute aggregates and plan the transaction rows: online money on the
	// booking's creation day (as before), on-site money on its recording day.
	// See planFinancialRows for the details; the enrichment (names, product)
	// stays here.
	summary, rows := planFinancialRows(decrypted, from, to)

	transactions := make([]domain.FinancialTransaction, 0, len(rows))
	for _, row := range rows {
		b := row.booking
		t := domain.FinancialTransaction{
			ID:            b.ID,
			SlotStartTime: b.SlotStartTime,
			PaidAt:        row.paidAt,
			AmountCents:   row.amountCents,
			RefundedCents: row.refundedCents,
			PaymentMode:   row.mode,
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
		Summary:      summary,
		Transactions: transactions,
	}, nil
}

// financialRow is one planned transaction row: which booking, which payment
// mode, the day the payment counts on and its amount. Rows are planned most
// recent payment first.
type financialRow struct {
	booking       *domain.Booking
	mode          domain.PaymentMode
	paidAt        time.Time
	amountCents   int
	refundedCents int
}

// planFinancialRows computes the summary totals and the transaction rows
// from decrypted bookings. Online money counts on the booking's creation day
// (domain.CollectedCents / RefundedCents, unchanged); an on-site payment
// counts on its own recording day (Europe/Paris), so a deposit paid online in
// one month and its balance paid at the practice in the next each show in
// their month. Disputed online payments stay tracked separately and out of
// gross/net; gross and net include on-site totals. A booking only appears as
// an online row when its creation day falls in [from, to) (the repository
// preselects that, re-checked here on the Paris day for consistency).
func planFinancialRows(bookings []*domain.Booking, from, to time.Time) (domain.FinancialSummary, []financialRow) {
	var grossRevenueCents, refundsCents, disputedCents, onSiteCents int
	var rows []financialRow

	for _, b := range bookings {
		if parisDayInRange(b.CreatedAt, from, to) {
			if b.PaymentStatus == domain.PaymentStatusDisputed {
				disputedCents += b.CollectedCents()
			} else {
				grossRevenueCents += b.CollectedCents()
				refundsCents += b.RefundedCents()
			}
			rows = append(rows, financialRow{
				booking:       b,
				mode:          domain.PaymentModeOnline,
				paidAt:        b.PaidAt(),
				amountCents:   b.CollectedCents(),
				refundedCents: b.RefundedCents(),
			})
		}
		if b.PaidOnSiteAt != nil && parisDayInRange(*b.PaidOnSiteAt, from, to) {
			amount := 0
			if b.AmountPaidOnSiteCents != nil {
				amount = *b.AmountPaidOnSiteCents
			}
			grossRevenueCents += amount
			onSiteCents += amount
			rows = append(rows, financialRow{
				booking:     b,
				mode:        domain.PaymentModeOnSite,
				paidAt:      *b.PaidOnSiteAt,
				amountCents: amount,
			})
		}
	}

	// Most recent payment first.
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].paidAt.After(rows[j].paidAt)
	})

	return domain.FinancialSummary{
		GrossRevenueCents: grossRevenueCents,
		RefundsCents:      refundsCents,
		DisputedCents:     disputedCents,
		OnSiteCents:       onSiteCents,
		NetRevenueCents:   grossRevenueCents - refundsCents,
	}, rows
}

// parisDayInRange reports whether t's Europe/Paris civil day falls in
// [from, to), both bounds read as civil days too (they are UTC midnights
// from the HTTP layer).
func parisDayInRange(t, from, to time.Time) bool {
	day := func(x time.Time) time.Time {
		p := x.In(domain.PracticeLocation)
		return time.Date(p.Year(), p.Month(), p.Day(), 0, 0, 0, 0, domain.PracticeLocation)
	}
	d := day(t)
	return !d.Before(day(from)) && d.Before(day(to))
}
