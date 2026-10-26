package couponPayment

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"

)

func (s *service) GetCoupon(ctx context.Context, stripeCouponID string) (*domain.Coupon, error) {
	stripeCoupon, err := s.V1Coupons.Retrieve(ctx, stripeCouponID, nil)
	if err != nil {
		return nil, err
	}

	return mapStripeCouponToDomainCoupon(stripeCoupon), nil
}