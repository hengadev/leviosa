package pricePayment

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"

	"github.com/stripe/stripe-go/v87"
)

func (s *service) CreatePrice(ctx context.Context, input domain.CreateStripePriceRequest) (*domain.PaymentPrice, error) {
	stripePrice, err := s.Client.V1Prices.Create(ctx, createPriceParams(input))
	if err != nil {
		return nil, err
	}

	return mapStripePriceToPaymentPrice(stripePrice), nil
}

// createPriceParams builds the Stripe request. A one-time price (the price
// of a booked session) has no recurring interval: Stripe accepts only day,
// week, month or year there, so sending "one_time" would fail the creation.
func createPriceParams(input domain.CreateStripePriceRequest) *stripe.PriceCreateParams {
	params := &stripe.PriceCreateParams{
		Product:    stripe.String(input.ProductID),
		UnitAmount: stripe.Int64(int64(input.Amount)),
		Currency:   stripe.String(input.Currency),
		Active:     stripe.Bool(input.Active),
	}

	if input.Interval != "" && domain.Interval(input.Interval) != domain.OneTime {
		params.Recurring = &stripe.PriceCreateRecurringParams{
			Interval: stripe.String(input.Interval),
		}
	}

	if input.Nickname != "" {
		params.Nickname = stripe.String(input.Nickname)
	}

	if len(input.Metadata) > 0 {
		params.Metadata = input.Metadata
	}

	return params
}

func mapStripePriceToPaymentPrice(stripePrice *stripe.Price) *domain.PaymentPrice {
	var interval string
	if stripePrice.Recurring != nil {
		interval = string(stripePrice.Recurring.Interval)
	}

	return &domain.PaymentPrice{
		ID:       stripePrice.ID,
		Product:  stripePrice.Product.ID,
		Amount:   int64(stripePrice.UnitAmount),
		Currency: string(stripePrice.Currency),
		Interval: interval,
		Active:   stripePrice.Active,
		Nickname: stripePrice.Nickname,
		Metadata: stripePrice.Metadata,
	}
}