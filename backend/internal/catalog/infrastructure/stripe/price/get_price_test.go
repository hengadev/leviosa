package pricePayment_test

import (
	"context"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPrice(t *testing.T) {
	ctx := context.Background()

	createMockProduct := func() string {
		return "prod_test123456789"
	}

	t.Run("successful price retrieval", func(t *testing.T) {
		// First create a price
		productID := createMockProduct()
		createReq := domain.CreateStripePriceRequest{
			ProductID: productID,
			Amount:    1999,
			Currency:  "USD",
			Interval:  "month",
			Active:    true,
			Nickname:  "Test Price for Get",
			Metadata: map[string]string{
				"test": "get_price",
			},
		}

		createdPrice, err := stripeService.CreatePrice(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, createdPrice)

		// Now retrieve the price
		retrievedPrice, err := stripeService.GetPrice(ctx, createdPrice.ID)

		require.NoError(t, err)
		assert.NotNil(t, retrievedPrice)
		assert.Equal(t, createdPrice.ID, retrievedPrice.ID)
		assert.Equal(t, createdPrice.Interval, retrievedPrice.Interval)
		assert.Equal(t, createdPrice.Active, retrievedPrice.Active)
		assert.Equal(t, createdPrice.Metadata, retrievedPrice.Metadata)
	})

	t.Run("retrieve price with minimal data", func(t *testing.T) {
		// Create a minimal price
		productID := createMockProduct()
		createReq := domain.CreateStripePriceRequest{
			ProductID: productID,
			Amount:    500,
			Currency:  "USD",
			Interval:  "month",
			Active:    true,
		}

		createdPrice, err := stripeService.CreatePrice(ctx, createReq)
		require.NoError(t, err)

		// Retrieve it
		retrievedPrice, err := stripeService.GetPrice(ctx, createdPrice.ID)

		require.NoError(t, err)
		assert.NotNil(t, retrievedPrice)
		assert.Equal(t, createdPrice.ID, retrievedPrice.ID)
		assert.Equal(t, "month", retrievedPrice.Interval)
		assert.True(t, retrievedPrice.Active)
	})

	t.Run("retrieve inactive price", func(t *testing.T) {
		// Create an inactive price
		productID := createMockProduct()
		createReq := domain.CreateStripePriceRequest{
			ProductID: productID,
			Amount:    750,
			Currency:  "EUR",
			Interval:  "year",
			Active:    false,
			Nickname:  "Inactive Price",
		}

		createdPrice, err := stripeService.CreatePrice(ctx, createReq)
		require.NoError(t, err)

		// Retrieve it
		retrievedPrice, err := stripeService.GetPrice(ctx, createdPrice.ID)

		require.NoError(t, err)
		assert.NotNil(t, retrievedPrice)
	})
}