package coupon

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

func (s *CouponService) UpdateCoupon(ctx context.Context, couponID string, req *domain.UpdateCouponRequest) error {
	if err := req.Valid(ctx); err != nil {
		return errs.NewInvalidValueErr(err.Error())
	}

	id, err := uuid.Parse(couponID)
	if err != nil {
		return errs.NewInvalidValueErr("invalid coupon ID format")
	}

	// Check if coupon exists
	_, err = s.repo.GetCouponByID(ctx, id)
	if err != nil {
		return fmt.Errorf("validate coupon: %w", err)
	}

	if err := s.repo.UpdateCoupon(ctx, id, req); err != nil {
		return fmt.Errorf("update coupon: %w", err)
	}

	return nil
}
