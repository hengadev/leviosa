package domain

import (
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFinancialSummaryResponse_ToCSV covers CSV serialization of an
// already-computed (and separately tested) aggregation result — it does not
// re-test the aggregation logic in application/booking/get_financial_summary.go,
// only the presentation transform per the financial-CSV-export issue.
func TestFinancialSummaryResponse_ToCSV(t *testing.T) {
	t.Run("serializes summary and transaction rows", func(t *testing.T) {
		txnID := uuid.New()
		slotStart := time.Date(2026, 7, 1, 9, 30, 0, 0, time.UTC)
		paidAt := time.Date(2026, 6, 20, 14, 5, 0, 0, time.UTC)

		resp := &FinancialSummaryResponse{
			Summary: FinancialSummary{
				GrossRevenueCents: 62300,
				RefundsCents:      2150,
				OnSiteCents:       10000,
				NetRevenueCents:   60150,
			},
			Transactions: []FinancialTransaction{
				{
					ID:            txnID,
					PaidAt:        paidAt,
					SlotStartTime: slotStart,
					ClientName:    "Marie Dupont",
					PartnerName:   "Dr. Martin",
					ProductName:   "Massage Relaxant 60min",
					AmountCents:   5000,
					RefundedCents: 0,
					PaymentMode:   PaymentModeOnline,
					PaymentStatus: PaymentStatusPaid,
					BookingStatus: BookingStatusCompleted,
				},
				{
					ID:            txnID,
					PaidAt:        slotStart,
					SlotStartTime: slotStart,
					ClientName:    "Marie Dupont",
					PartnerName:   "Dr. Martin",
					ProductName:   "Massage Relaxant 60min",
					AmountCents:   10000,
					RefundedCents: 0,
					PaymentMode:   PaymentModeOnSite,
					PaymentStatus: PaymentStatusPaid,
					BookingStatus: BookingStatusCompleted,
				},
			},
		}

		out, err := resp.ToCSV()
		require.NoError(t, err)

		reader := csv.NewReader(strings.NewReader(string(out)))
		// The summary block uses a variable number of fields per row (the
		// blank separator row has none), so relax the reader's field-count
		// check to parse the whole file.
		reader.FieldsPerRecord = -1
		records, err := reader.ReadAll()
		require.NoError(t, err)

		// The blank separator line between the summary and transaction
		// sections is ignored by encoding/csv's reader, so it doesn't
		// produce its own record.
		require.Len(t, records, 8)

		assert.Equal(t, []string{"section", "summary"}, records[0])
		assert.Equal(t, []string{"gross_revenue_cents", "62300"}, records[1])
		assert.Equal(t, []string{"refunds_cents", "2150"}, records[2])
		assert.Equal(t, []string{"on_site_cents", "10000"}, records[3])
		assert.Equal(t, []string{"net_revenue_cents", "60150"}, records[4])

		assert.Equal(t, []string{
			"id", "paid_at", "slot_start_time", "client_name", "partner_name",
			"product_name", "amount_cents", "refunded_cents", "payment_status", "booking_status", "mode",
		}, records[5])

		assert.Equal(t, []string{
			txnID.String(),
			"2026-06-20T14:05:00Z",
			"2026-07-01T09:30:00Z",
			"Marie Dupont",
			"Dr. Martin",
			"Massage Relaxant 60min",
			"5000",
			"0",
			"paid",
			"completed",
			"En ligne",
		}, records[6])
		assert.Equal(t, []string{
			txnID.String(),
			"2026-07-01T09:30:00Z",
			"2026-07-01T09:30:00Z",
			"Marie Dupont",
			"Dr. Martin",
			"Massage Relaxant 60min",
			"10000",
			"0",
			"paid",
			"completed",
			"Sur place",
		}, records[7])
	})

	t.Run("handles no transactions without error", func(t *testing.T) {
		resp := &FinancialSummaryResponse{
			Summary: FinancialSummary{},
		}

		out, err := resp.ToCSV()
		require.NoError(t, err)

		reader := csv.NewReader(strings.NewReader(string(out)))
		reader.FieldsPerRecord = -1
		records, err := reader.ReadAll()
		require.NoError(t, err)

		// Summary block (5 rows) + transaction header, no data rows (the
		// blank separator line is ignored by encoding/csv's reader).
		require.Len(t, records, 6)
		assert.Equal(t, []string{
			"id", "paid_at", "slot_start_time", "client_name", "partner_name",
			"product_name", "amount_cents", "refunded_cents", "payment_status", "booking_status", "mode",
		}, records[5])
	})

	t.Run("nil response produces a valid empty CSV", func(t *testing.T) {
		var resp *FinancialSummaryResponse

		out, err := resp.ToCSV()
		require.NoError(t, err)
		assert.NotEmpty(t, out)
	})

	t.Run("multiple transactions preserve order and each row's data", func(t *testing.T) {
		id1, id2 := uuid.New(), uuid.New()
		resp := &FinancialSummaryResponse{
			Summary: FinancialSummary{GrossRevenueCents: 11000, RefundsCents: 5000, NetRevenueCents: 6000},
			Transactions: []FinancialTransaction{
				{
					ID:            id1,
					SlotStartTime: time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC),
					ClientName:    "Jean Durand",
					PartnerName:   "Dr. Martin",
					ProductName:   "Soin du Dos",
					AmountCents:   6000,
					PaymentMode:   PaymentModeOnline,
					PaymentStatus: PaymentStatusPaid,
					BookingStatus: BookingStatusCompleted,
				},
				{
					ID:            id2,
					SlotStartTime: time.Date(2026, 7, 3, 11, 0, 0, 0, time.UTC),
					ClientName:    "Claire Bernard",
					PartnerName:   "Dr. Dupont",
					ProductName:   "Drainage Lymphatique",
					AmountCents:   5000,
					RefundedCents: 5000,
					PaymentMode:   PaymentModeOnSite,
					PaymentStatus: PaymentStatusRefunded,
					BookingStatus: BookingStatusCancelled,
				},
			},
		}

		out, err := resp.ToCSV()
		require.NoError(t, err)

		reader := csv.NewReader(strings.NewReader(string(out)))
		reader.FieldsPerRecord = -1
		records, err := reader.ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 8)
		assert.Equal(t, id1.String(), records[6][0])
		assert.Equal(t, "paid", records[6][8])
		assert.Equal(t, "En ligne", records[6][10])
		assert.Equal(t, id2.String(), records[7][0])
		assert.Equal(t, "5000", records[7][7], "the refunded amount")
		assert.Equal(t, "refunded", records[7][8])
		assert.Equal(t, "Sur place", records[7][10])
	})
}
