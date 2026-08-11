package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newValidTestPrice() Price {
	return Price{
		Amount:   2000,
		Currency: "EUR",
		Interval: Month,
	}
}

func TestPrice_Valid_TaxRatePercent(t *testing.T) {
	ctx := context.Background()

	t.Run("nil tax rate is valid (default exempt)", func(t *testing.T) {
		p := newValidTestPrice()
		p.TaxRatePercent = nil

		err := p.Valid(ctx)

		assert.NoError(t, err)
	})

	t.Run("tax rate within 0-100 is valid", func(t *testing.T) {
		for _, rate := range []float64{0, 5.5, 20, 100} {
			r := rate
			p := newValidTestPrice()
			p.TaxRatePercent = &r

			err := p.Valid(ctx)

			assert.NoError(t, err, "rate %v should be valid", rate)
		}
	})

	t.Run("negative tax rate is invalid", func(t *testing.T) {
		rate := -1.0
		p := newValidTestPrice()
		p.TaxRatePercent = &rate

		err := p.Valid(ctx)

		assert.Error(t, err)
	})

	t.Run("tax rate above 100 is invalid", func(t *testing.T) {
		rate := 100.01
		p := newValidTestPrice()
		p.TaxRatePercent = &rate

		err := p.Valid(ctx)

		assert.Error(t, err)
	})
}

func TestPrice_TaxLine_ExemptPath(t *testing.T) {
	p := newValidTestPrice()
	p.Amount = 10000
	p.TaxRatePercent = nil

	t.Run("no VAT number configured", func(t *testing.T) {
		line := p.TaxLine("")

		assert.True(t, line.Exempt)
		assert.Nil(t, line.RatePercent)
		assert.Equal(t, 0, line.TaxAmount)
		assert.Equal(t, p.Amount, line.TotalAmount)
		assert.Equal(t, VATExemptionNote, line.Note)
	})

	t.Run("VAT number configured does not change the exemption note", func(t *testing.T) {
		// Exemption is driven purely by whether a rate is set on the price;
		// the exemption note stays the standard franchise-en-base mention
		// regardless of whether a VAT number happens to be configured.
		line := p.TaxLine("FR12345678901")

		assert.True(t, line.Exempt)
		assert.Equal(t, VATExemptionNote, line.Note)
	})
}

func TestPrice_TaxLine_RateSetPath(t *testing.T) {
	p := newValidTestPrice()
	p.Amount = 10000 // 100.00 EUR
	rate := 20.0
	p.TaxRatePercent = &rate

	t.Run("computes tax line without a VAT number", func(t *testing.T) {
		line := p.TaxLine("")

		assert.False(t, line.Exempt)
		if assert.NotNil(t, line.RatePercent) {
			assert.Equal(t, 20.0, *line.RatePercent)
		}
		assert.Equal(t, 2000, line.TaxAmount)    // 20% of 10000
		assert.Equal(t, 12000, line.TotalAmount) // 10000 + 2000
		assert.Contains(t, line.Note, "20.00%")
		assert.NotContains(t, line.Note, "VAT no.")
	})

	t.Run("includes the VAT number in the note when configured", func(t *testing.T) {
		line := p.TaxLine("FR12345678901")

		assert.False(t, line.Exempt)
		assert.Contains(t, line.Note, "FR12345678901")
	})

	t.Run("rounds computed tax amount to the nearest cent", func(t *testing.T) {
		oddRate := 5.5
		odd := newValidTestPrice()
		odd.Amount = 999
		odd.TaxRatePercent = &oddRate

		line := odd.TaxLine("")

		// 999 * 5.5 / 100 = 54.945 -> rounds to 55
		assert.Equal(t, 55, line.TaxAmount)
		assert.Equal(t, 999+55, line.TotalAmount)
	})

	t.Run("zero rate produces no tax but is not treated as exempt", func(t *testing.T) {
		zero := 0.0
		p2 := newValidTestPrice()
		p2.Amount = 5000
		p2.TaxRatePercent = &zero

		line := p2.TaxLine("")

		assert.False(t, line.Exempt)
		assert.Equal(t, 0, line.TaxAmount)
		assert.Equal(t, 5000, line.TotalAmount)
		assert.NotEqual(t, VATExemptionNote, line.Note)
	})
}
