package stripeService

import (
	"context"
	"os"

	"github.com/GaryHY/leviosa/internal/domain"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
)

// TODO: add in the checkoutsessionparams the field:
// - customer: so that is easier for them to pay next time
// - metadata: I think I can use that for the webhook after that.
// So here, I will send the userID and the eventID so that I can use that easily in the handler for the wh

func (s *Service) CreateCheckoutSession(ctx context.Context, priceID string, quantity int64) (string, error) {
	frontendServer := os.Getenv("FRONTEND_ORIGIN")
	// NOTE: I can use metadata if I want but I do not know why I would use that, maybe with the product type
	// metadata := map[string]string{
	// 	"eventID": eventID,
	// 	"userID":  userID,
	// 	"spot":    spot,
	// }
	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(quantity),
			},
		},
		Mode:         stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:   stripe.String(frontendServer + "/app/checkout/success/"),
		CancelURL:    stripe.String(frontendServer + "/app/checkout/cancel/"),
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{Enabled: stripe.Bool(true)},
		// Metadata:     metadata,
	}
	checkoutSession, err := session.New(params)
	if err != nil {
		return "", domain.NewNotCreatedErr(err)
	}
	return checkoutSession.URL, nil
}
