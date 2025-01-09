package stripeService

import (
	"context"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/price"
)

func (s *Service) CreatePrice(ctx context.Context, productID string, priceValue int64) (string, error) {
	price_params := &stripe.PriceParams{
		Currency:   stripe.String(string(stripe.CurrencyEUR)),
		Product:    stripe.String(productID),
		UnitAmount: stripe.Int64(priceValue),
	}
	price, err := price.New(price_params)
	if err != nil {
		return "", domain.NewNotCreatedErr(err)
	}
	return price.ID, nil
}
