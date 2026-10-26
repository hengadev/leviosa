package promotionCodePayment

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"

	"github.com/stripe/stripe-go/v82"
)

func (s *service) UpdatePromotionCode(ctx context.Context, stripePromotionID string, req *domain.UpdatePromotionCodeRequest) (*domain.PromotionCode, error) {
	params := &stripe.PromotionCodeUpdateParams{}

	if req.Active != nil {
		params.Active = stripe.Bool(*req.Active)
	}

	if req.Metadata != nil {
		params.Metadata = req.Metadata
	}

	stripePromotionCode, err := s.V1PromotionCodes.Update(ctx, stripePromotionID, params)
	if err != nil {
		return nil, err
	}

	return mapStripePromotionCodeToDomainPromotionCode(stripePromotionCode), nil
}