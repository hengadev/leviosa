package pricePayment

import (
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePriceParams(t *testing.T) {
	base := domain.CreateStripePriceRequest{ProductID: "prod_1", Amount: 6000, Currency: "eur", Active: true}

	t.Run("a one-time price omits the recurring interval", func(t *testing.T) {
		input := base
		input.Interval = string(domain.OneTime)
		params := createPriceParams(input)
		assert.Nil(t, params.Recurring)
		assert.Equal(t, int64(6000), *params.UnitAmount)
	})

	t.Run("a price without an interval is one-time", func(t *testing.T) {
		assert.Nil(t, createPriceParams(base).Recurring)
	})

	for _, interval := range []domain.Interval{domain.Month, domain.Year} {
		t.Run("a "+string(interval)+" price recurs", func(t *testing.T) {
			input := base
			input.Interval = string(interval)
			params := createPriceParams(input)
			require.NotNil(t, params.Recurring)
			assert.Equal(t, string(interval), *params.Recurring.Interval)
		})
	}
}
