package promotionCodePayment

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"

	"github.com/stripe/stripe-go/v82"
)

func (s *service) ListPromotionCodes(ctx context.Context, stripeCouponID string) ([]*domain.PromotionCode, error) {
	params := &stripe.PromotionCodeListParams{
		Coupon: stripe.String(stripeCouponID),
	}

	var promotionCodes []*domain.PromotionCode

	for stripePromotionCode, err := range s.V1PromotionCodes.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		promotionCodes = append(promotionCodes, mapStripePromotionCodeToDomainPromotionCode(stripePromotionCode))
	}

	return promotionCodes, nil
}