package couponPayment_test

import (
	"context"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCoupon(t *testing.T) {
	ctx := context.Background()

	t.Run("successful coupon retrieval", func(t *testing.T) {
		// First create a coupon
		percentOff := 20.0
		maxRedemptions := 50

		createReq := &domain.CreateCouponRequest{
			Name:           "Test Coupon for Get",
			PercentOff:     &percentOff,
			Duration:       "once",
			MaxRedemptions: &maxRedemptions,
			Metadata: map[string]string{
				"test": "get_coupon",
			},
		}

		createdCoupon, err := stripeService.CreateCoupon(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, createdCoupon)

		// Now retrieve the coupon
		retrievedCoupon, err := stripeService.GetCoupon(ctx, createdCoupon.StripeCouponID)

		require.NoError(t, err)
		assert.NotNil(t, retrievedCoupon)
		assert.Equal(t, createdCoupon.StripeCouponID, retrievedCoupon.StripeCouponID)
		assert.Equal(t, createdCoupon.IsValid, retrievedCoupon.IsValid)
		assert.Equal(t, createdCoupon.Metadata, retrievedCoupon.Metadata)
		assert.Equal(t, createdCoupon.TimesRedeemed, retrievedCoupon.TimesRedeemed)
	})

	t.Run("retrieve amount-off coupon", func(t *testing.T) {
		// Create an amount-off coupon
		amountOff := 1000 // $10.00
		currency := "USD"

		createReq := &domain.CreateCouponRequest{
			Name:      "Amount Off Test Coupon",
			AmountOff: &amountOff,
			Currency:  &currency,
			Duration:  "once",
			Metadata: map[string]string{
				"type": "amount_off",
			},
		}

		createdCoupon, err := stripeService.CreateCoupon(ctx, createReq)
		require.NoError(t, err)

		// Retrieve it
		retrievedCoupon, err := stripeService.GetCoupon(ctx, createdCoupon.StripeCouponID)

		require.NoError(t, err)
		assert.NotNil(t, retrievedCoupon)
		assert.Equal(t, createdCoupon.StripeCouponID, retrievedCoupon.StripeCouponID)
		assert.Equal(t, createdCoupon.Metadata, retrievedCoupon.Metadata)
	})

	t.Run("retrieve repeating coupon", func(t *testing.T) {
		// Create a repeating coupon
		percentOff := 15.0
		durationInMonths := 6

		createReq := &domain.CreateCouponRequest{
			Name:             "Repeating Test Coupon",
			PercentOff:       &percentOff,
			Duration:         "repeating",
			DurationInMonths: &durationInMonths,
		}

		createdCoupon, err := stripeService.CreateCoupon(ctx, createReq)
		require.NoError(t, err)

		// Retrieve it
		retrievedCoupon, err := stripeService.GetCoupon(ctx, createdCoupon.StripeCouponID)

		require.NoError(t, err)
		assert.NotNil(t, retrievedCoupon)
	})

	t.Run("retrieve forever coupon", func(t *testing.T) {
		// Create a forever coupon
		percentOff := 5.0

		createReq := &domain.CreateCouponRequest{
			Name:       "Forever Test Coupon",
			PercentOff: &percentOff,
			Duration:   "forever",
		}

		createdCoupon, err := stripeService.CreateCoupon(ctx, createReq)
		require.NoError(t, err)

		// Retrieve it
		retrievedCoupon, err := stripeService.GetCoupon(ctx, createdCoupon.StripeCouponID)

		require.NoError(t, err)
		assert.NotNil(t, retrievedCoupon)
		assert.Equal(t, domain.CouponDuration("forever"), retrievedCoupon.Duration)
	})

	t.Run("retrieve coupon with minimal data", func(t *testing.T) {
		// Create a minimal coupon
		percentOff := 25.0

		createReq := &domain.CreateCouponRequest{
			Name:       "Minimal Coupon for Get",
			PercentOff: &percentOff,
			Duration:   "once",
		}

		createdCoupon, err := stripeService.CreateCoupon(ctx, createReq)
		require.NoError(t, err)

		// Retrieve it
		retrievedCoupon, err := stripeService.GetCoupon(ctx, createdCoupon.StripeCouponID)

		require.NoError(t, err)
		assert.NotNil(t, retrievedCoupon)
		assert.True(t, retrievedCoupon.IsValid)
		assert.Equal(t, 0, retrievedCoupon.TimesRedeemed)
	})
}