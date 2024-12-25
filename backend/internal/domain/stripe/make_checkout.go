package stripeService

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
)

// TODO: add in the checkoutsessionparams the field:
// - customer: so that is easier for them to pay next time
// - metadata: I think I can use that for the webhook after that.
// So here, I will send the userID and the eventID so that I can use that easily in the handler for the wh

func (s *Service) CreateCheckoutSession(ctx context.Context, domain, priceID, eventID, userID, spot string) (string, error) {
	metadata := map[string]string{
		"eventID": eventID,
		"userID":  userID,
		"spot":    spot,
	}
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
		Metadata:     metadata,
	}
	checkoutSession, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("Error with creating the new session Stripe : %v", err)
	}
	return checkoutSession.URL, nil
}
