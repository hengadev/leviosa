package promotionCodePayment

import (
	"context"

	"github.com/stripe/stripe-go/v82"
)

func (s *service) DeletePromotionCode(ctx context.Context, stripePromotionID string) error {
	// Note: Stripe doesn't have a delete operation for promotion codes
	// Instead, we deactivate the promotion code
	params := &stripe.PromotionCodeUpdateParams{
		Active: stripe.Bool(false),
	}

	_, err := s.V1PromotionCodes.Update(ctx, stripePromotionID, params)
	return err
}