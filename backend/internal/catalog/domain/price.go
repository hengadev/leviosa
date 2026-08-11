package domain

import (
	"context"
	"fmt"
	"math"
	"regexp" // For currency code validation
	"time"

	"github.com/google/uuid"
	"github.com/hengadev/errsx"
)

type Price struct {
	ID            uuid.UUID
	ProductID     uuid.UUID `json:"productId"`     // FK to your internal product
	StripePriceID string    `json:"stripePriceId"` // Price ID from Stripe
	Amount        int       `json:"amount"`        // in cents
	Currency      string    `json:"currency"`      // "usd", "eur", etc.
	Interval      Interval  `json:"interval"`      // "month", "year", "one_time"
	IsActive      bool      `json:"isActive"`      // Track which one is current
	// TaxRatePercent is an optional VAT/tax rate expressed as a percentage
	// (e.g. 20 for 20%). Nil means exempt/unset — the safe default for a
	// business that may operate under the French "franchise en base de TVA"
	// regime, which is legally required to charge no VAT. This is never
	// assumed or defaulted to a non-null value anywhere in the system; it
	// must be explicitly set. Stored and applied at display/calculation
	// time only — not synced to Stripe.
	TaxRatePercent *float64  `json:"taxRatePercent,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

var currencyRegex = regexp.MustCompile("^[A-Z]{3}$")

func (p Price) Valid(ctx context.Context) error {
	var errs errsx.Map
	// Amount validation
	if p.Amount <= 0 {
		errs.Set("amount", "must be a positive value")
	}

	// Currency validation (3-letter uppercase ISO code)
	if !currencyRegex.MatchString(p.Currency) {
		errs.Set("currency", "must be a 3-letter uppercase ISO currency code (e.g., 'USD', 'EUR')")
	}

	// Interval validation (must be one of the defined ENUM values)
	if !p.Interval.IsValid() {
		errs.Set("interval", fmt.Sprintf("Invalid interval type: '%s'. Must be 'month', 'year', or 'one_time'.", p.Interval))
	}

	// TaxRatePercent validation: optional, but when present must be a sane percentage.
	if p.TaxRatePercent != nil && (*p.TaxRatePercent < 0 || *p.TaxRatePercent > 100) {
		errs.Set("taxRatePercent", "must be between 0 and 100 when set")
	}

	return errs.AsError()
}

// VATExemptionNote is the standard French mention for businesses operating
// under the "franchise en base de TVA" regime (art. 293 B du CGI), which are
// legally required to charge no VAT and state this explicitly instead of a
// tax line.
const VATExemptionNote = "TVA non applicable, art. 293 B du CGI"

// TaxLine is the display/calculation result for a Price's tax treatment,
// suitable for rendering on quotes, invoices, or receipts.
type TaxLine struct {
	// Exempt is true when no tax rate is set on the price (the default).
	Exempt bool `json:"exempt"`
	// RatePercent is the tax rate applied, only set when Exempt is false.
	RatePercent *float64 `json:"ratePercent,omitempty"`
	// TaxAmount is the computed tax in cents (0 when Exempt).
	TaxAmount int `json:"taxAmount"`
	// TotalAmount is Price.Amount + TaxAmount, in cents.
	TotalAmount int `json:"totalAmount"`
	// Note is the human-readable line to display: either the computed tax
	// description (rate set) or the VAT exemption note (rate null).
	Note string `json:"note"`
}

// TaxLine computes the display/calculation info for this price's tax
// treatment. vatNumber is the business's VAT number from settings (issue
// 001, "legal identity fields"), if any has been configured; it is passed in
// rather than looked up here to keep the catalog domain decoupled from the
// settings domain. An empty vatNumber is handled gracefully.
//
//   - When TaxRatePercent is set, a computed tax line is returned, including
//     the VAT number in the note when one is available (a business charging
//     VAT is expected to display its VAT number).
//   - When TaxRatePercent is nil, the price is treated as exempt and the
//     standard franchise-en-base exemption note is returned instead.
func (p Price) TaxLine(vatNumber string) TaxLine {
	if p.TaxRatePercent == nil {
		return TaxLine{
			Exempt:      true,
			TaxAmount:   0,
			TotalAmount: p.Amount,
			Note:        VATExemptionNote,
		}
	}

	rate := *p.TaxRatePercent
	taxAmount := int(math.Round(float64(p.Amount) * rate / 100))

	note := fmt.Sprintf("VAT (%.2f%%)", rate)
	if vatNumber != "" {
		note = fmt.Sprintf("VAT (%.2f%%) — VAT no. %s", rate, vatNumber)
	}

	return TaxLine{
		Exempt:      false,
		RatePercent: &rate,
		TaxAmount:   taxAmount,
		TotalAmount: p.Amount + taxAmount,
		Note:        note,
	}
}

type Interval string

const (
	OneTime Interval = "one_time"
	Month   Interval = "month"
	Year    Interval = "year"
)

// IsValid checks if the AvailabilityType is one of the defined constants.
func (it Interval) IsValid() bool {
	switch it {
	case Month, Year, OneTime:
		return true
	default:
		return false
	}
}
