package domain

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FinancialSummary holds the aggregate KPIs for a date range.
//
// DisputedCents is tracked separately from GrossRevenueCents/NetRevenueCents: a disputed
// payment's final outcome (won/lost) is not yet known, so it is excluded from both until
// the dispute closes (at which point the booking's payment status resolves back to paid or
// refunded, per ADR-0007 — booking-derived, no Stripe Balance/Payout integration).
type FinancialSummary struct {
	GrossRevenueCents int `json:"gross_revenue_cents"`
	RefundsCents      int `json:"refunds_cents"`
	DisputedCents     int `json:"disputed_cents"`
	NetRevenueCents   int `json:"net_revenue_cents"`
}

// FinancialTransaction is a single row in the financial transactions list.
type FinancialTransaction struct {
	ID              uuid.UUID     `json:"id"`
	SlotStartTime   time.Time     `json:"slot_start_time"`
	ClientName      string        `json:"client_name"`
	PartnerName     string        `json:"partner_name"`
	ProductName     string        `json:"product_name"`
	AmountCents     int           `json:"amount_cents"`
	PaymentStatus   PaymentStatus `json:"payment_status"`
	BookingStatus   BookingStatus `json:"booking_status"`
}

// FinancialSummaryResponse is the response shape for GET /admin/bookings/financial-summary.
type FinancialSummaryResponse struct {
	Summary      FinancialSummary      `json:"summary"`
	Transactions []FinancialTransaction `json:"transactions"`
}

// ToCSV serializes an already-computed FinancialSummaryResponse into CSV bytes.
// It performs no aggregation of its own — it is a pure presentation-layer
// transform over the same result the JSON endpoint returns, per ADR-0007
// (financial reporting stays booking-derived; this only changes the response
// format, not how the numbers are computed).
//
// The output has two sections: a summary block with the period KPIs (gross
// revenue, refunds, net revenue) that mirror the compta/finances UI cards,
// followed by a blank line and the transaction detail rows. Amounts are
// written in cents (integers) to avoid floating-point rounding when the
// comptable imports the file into accounting software.
func (r *FinancialSummaryResponse) ToCSV() ([]byte, error) {
	if r == nil {
		r = &FinancialSummaryResponse{}
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	summaryRows := [][]string{
		{"section", "summary"},
		{"gross_revenue_cents", fmt.Sprintf("%d", r.Summary.GrossRevenueCents)},
		{"refunds_cents", fmt.Sprintf("%d", r.Summary.RefundsCents)},
		{"net_revenue_cents", fmt.Sprintf("%d", r.Summary.NetRevenueCents)},
		{},
	}
	if err := w.WriteAll(summaryRows); err != nil {
		return nil, fmt.Errorf("write CSV summary rows: %w", err)
	}

	header := []string{
		"id",
		"slot_start_time",
		"client_name",
		"partner_name",
		"product_name",
		"amount_cents",
		"payment_status",
		"booking_status",
	}
	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("write CSV transaction header: %w", err)
	}

	for _, t := range r.Transactions {
		row := []string{
			t.ID.String(),
			t.SlotStartTime.UTC().Format(time.RFC3339),
			t.ClientName,
			t.PartnerName,
			t.ProductName,
			fmt.Sprintf("%d", t.AmountCents),
			string(t.PaymentStatus),
			string(t.BookingStatus),
		}
		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("write CSV transaction row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush CSV writer: %w", err)
	}

	return buf.Bytes(), nil
}
