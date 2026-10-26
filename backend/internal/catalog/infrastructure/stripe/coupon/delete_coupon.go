package couponPayment

import (
	"context"

)

func (s *service) DeleteCoupon(ctx context.Context, stripeCouponID string) error {
	_, err := s.V1Coupons.Delete(ctx, stripeCouponID, nil)
	return err
}