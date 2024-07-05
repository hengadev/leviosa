package checkout

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
)

func (s *Service) CreateCheckoutSession(ctx context.Context, domain, priceID string) (string, error) {
	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:         stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:   stripe.String(domain + "app/checkout/success/"),
		CancelURL:    stripe.String(domain + "app/checkout/cancel/"),
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{Enabled: stripe.Bool(true)},
	}
	checkoutSession, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("Error with creating the new session Stripe : %v", err)
	}
	return checkoutSession.URL, nil
}
