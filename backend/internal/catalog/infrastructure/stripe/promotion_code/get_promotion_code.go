package promotionCodePayment

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"

)

func (s *service) GetPromotionCode(ctx context.Context, stripePromotionID string) (*domain.PromotionCode, error) {
	stripePromotionCode, err := s.V1PromotionCodes.Retrieve(ctx, stripePromotionID, nil)
	if err != nil {
		return nil, err
	}

	return mapStripePromotionCodeToDomainPromotionCode(stripePromotionCode), nil
}