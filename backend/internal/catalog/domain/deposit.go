package domain

import (
	"encoding/json"
	"fmt"
	"math"
)

// Deposit is the part of a product's price paid online when booking (French:
// *acompte*, never *arrhes*) — issue 013, ADR-0018. A product without a
// deposit has its full price paid online; with one, only the deposit goes
// through Leviosa and the balance is paid at the practice.
//
// A deposit is either a fixed amount (Value in cents) or a percentage of the
// price (Value in whole percentage points).
type Deposit struct {
	// Type selects how Value is interpreted: fixed cents or a percentage of
	// the price.
	Type DepositType `json:"type"`
	// Value is the deposit amount in cents when Type is DepositFixed, or the
	// percentage of the price when Type is DepositPercent.
	Value int `json:"value"`
}

// DepositType discriminates the two deposit shapes.
type DepositType string

const (
	// DepositFixed: Value is an amount in cents, between 0 and the price.
	DepositFixed DepositType = "fixed"
	// DepositPercent: Value is a percentage of the price, in (0, 100].
	DepositPercent DepositType = "percent"
)

// IsValid checks the DepositType is one of the defined constants.
func (t DepositType) IsValid() bool {
	switch t {
	case DepositFixed, DepositPercent:
		return true
	default:
		return false
	}
}

// Validate checks a deposit's structural constraints: a known type, a
// non-negative fixed amount, and a percentage in (0, 100]. The
// price-relative rule for fixed deposits (amount <= the product's price) is
// ValidateAgainstPrice: the Product entity alone does not know its price,
// which lives in separate Stripe-backed price records.
func (d *Deposit) Validate() error {
	if d == nil {
		return nil
	}
	switch d.Type {
	case DepositFixed:
		if d.Value < 0 {
			return fmt.Errorf("fixed deposit amount cannot be negative, got %d cents", d.Value)
		}
	case DepositPercent:
		if d.Value <= 0 || d.Value > 100 {
			return fmt.Errorf("deposit percentage must be in (0, 100], got %d", d.Value)
		}
	default:
		return fmt.Errorf("invalid deposit type: '%s'. Must be 'fixed' or 'percent'.", d.Type)
	}
	return nil
}

// ValidateAgainstPrice checks the deposit's structural constraints and, for
// a fixed deposit, that it does not exceed the price (priceCents). It is the
// full rule, applied wherever the deposit and the price are both known:
// creating a product with its price, changing a product's deposit while it
// has active prices, and adding a new price to a product that has a
// deposit. A Stripe price's amount is immutable, so a price "change" is a
// new price and goes through the last of these checks.
func (d *Deposit) ValidateAgainstPrice(priceCents int) error {
	if d == nil {
		return nil
	}
	_, err := AmountDueOnline(d, priceCents)
	return err
}

// AmountDueOnline returns the amount a client pays online when booking: the
// product's deposit when one is configured, otherwise the full price
// (ADR-0018). A percentage deposit is rounded to the cent.
//
// It is pure, so booking creation (issue 014) and any other caller resolve
// the online amount from the same rule the catalog domain validated. A fixed
// deposit above the price is rejected: the online amount can never exceed
// the price.
func AmountDueOnline(deposit *Deposit, priceCents int) (int, error) {
	if deposit == nil {
		return priceCents, nil
	}
	switch deposit.Type {
	case DepositFixed:
		if deposit.Value < 0 {
			return 0, fmt.Errorf("fixed deposit amount cannot be negative, got %d cents", deposit.Value)
		}
		if deposit.Value > priceCents {
			return 0, fmt.Errorf("fixed deposit of %d cents exceeds the price of %d cents", deposit.Value, priceCents)
		}
		return deposit.Value, nil
	case DepositPercent:
		if deposit.Value <= 0 || deposit.Value > 100 {
			return 0, fmt.Errorf("deposit percentage must be in (0, 100], got %d", deposit.Value)
		}
		return int(math.Round(float64(priceCents) * float64(deposit.Value) / 100)), nil
	default:
		return 0, fmt.Errorf("invalid deposit type: '%s'. Must be 'fixed' or 'percent'.", deposit.Type)
	}
}

// DepositPatch carries a deposit through a PATCH update. JSON objects set or
// replace the deposit, JSON null removes it, and an absent field leaves it
// unchanged — plain *Deposit cannot tell "null" from "absent".
type DepositPatch struct {
	// Present reports whether the "deposit" field appeared in the JSON.
	Present bool
	// Value is the deposit to set, or nil to remove the deposit.
	Value *Deposit
}

// UnmarshalJSON records the field's presence so the update layer can
// distinguish set, clear and leave-unchanged.
func (p *DepositPatch) UnmarshalJSON(data []byte) error {
	p.Present = true
	if string(data) == "null" {
		p.Value = nil
		return nil
	}
	var d Deposit
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	p.Value = &d
	return nil
}

// MarshalJSON writes the deposit to set, or null to remove it. Paired with
// the `omitzero` tag, an untouched patch is left out entirely, so a request
// marshalled by a Go client round-trips with the same meaning.
func (p DepositPatch) MarshalJSON() ([]byte, error) {
	if p.Value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(p.Value)
}
