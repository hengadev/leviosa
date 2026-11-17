package booking

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/google/uuid"
)

const earningsMaxFetch = 10000

func (s *BookingService) GetPartnerEarnings(ctx context.Context, partnerID uuid.UUID) (*domain.EarningsSummary, error) {
	// Earnings count appointments only (ADR-0018).
	filter := ports.BookingFilter{Limit: earningsMaxFetch, ExcludeStatus: domain.NonAppointmentStatuses()}
	bookingsEncx, err := s.bookingRepo.GetByPartnerID(ctx, partnerID, filter)
	if err != nil {
		return nil, fmt.Errorf("get partner bookings: %w", err)
	}

	if len(bookingsEncx) == earningsMaxFetch {
		slog.WarnContext(ctx, "partner earnings fetch hit safety cap; results may be incomplete",
			"cap", earningsMaxFetch, "partner_id", partnerID)
	}

	bookings := make([]*domain.Booking, 0, len(bookingsEncx))
	for _, bookingEncx := range bookingsEncx {
		booking, err := domain.DecryptBookingEncx(ctx, s.crypto, bookingEncx)
		if err != nil {
			return nil, fmt.Errorf("decrypt booking %s: %w", bookingEncx.ID, err)
		}
		bookings = append(bookings, booking)
	}

	// Build product name cache to avoid N+1 calls to the catalog service.
	productNames := make(map[uuid.UUID]string)
	if s.productService != nil {
		for _, booking := range bookings {
			if _, ok := productNames[booking.ProductID]; !ok {
				productNames[booking.ProductID] = s.resolveProductName(ctx, booking.ProductID)
			}
		}
	}

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	lastMonthDate := now.AddDate(0, -1, 0)
	lastYear, lastMonth, _ := lastMonthDate.Date()
	sevenDaysAgo := now.AddDate(0, 0, -7)

	summary := &domain.EarningsSummary{
		CurrentMonthCents: 0,
		LastMonthCents:    0,
		PendingCents:      0,
		NextPayoutCents:   0,
		NextPayoutDate:    nextMonday(now).Format(time.RFC3339),
		Transactions:      []domain.Transaction{},
	}

	// Earnings are the money collected, on the day it was paid (see
	// domain.CollectedCents): a deposit paid today counts today, whatever
	// the date of the session.
	for _, booking := range bookings {
		paidAt := booking.PaidAt()
		paidYear, paidMonth, _ := paidAt.Date()
		net := booking.NetCollectedCents()

		if paidYear == currentYear && paidMonth == currentMonth {
			summary.CurrentMonthCents += net
		}
		if paidYear == lastYear && paidMonth == lastMonth {
			summary.LastMonthCents += net
		}
		if !paidAt.Before(sevenDaysAgo) {
			summary.NextPayoutCents += net
		}

		if booking.PaymentStatus == domain.PaymentStatusPending {
			summary.PendingCents += booking.TotalPriceCents
		}

		productName, ok := productNames[booking.ProductID]
		if !ok {
			productName = "Produit inconnu"
		}

		// What was paid online; the price still due for an unpaid booking.
		amount := booking.CollectedCents()
		if amount == 0 {
			amount = booking.TotalPriceCents
		}

		summary.Transactions = append(summary.Transactions, domain.Transaction{
			ID:            booking.ID,
			SlotStartTime: booking.SlotStartTime.Format(time.RFC3339),
			PaidAt:        paidAt.Format(time.RFC3339),
			ProductID:     booking.ProductID,
			ProductName:   productName,
			AmountCents:   amount,
			NetCents:      net,
			PaymentStatus: booking.PaymentStatus,
			BookingStatus: booking.Status,
		})
	}

	return summary, nil
}

func nextMonday(t time.Time) time.Time {
	daysUntilMonday := (int(time.Monday) - int(t.Weekday()) + 7) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	return t.AddDate(0, 0, daysUntilMonday)
}
