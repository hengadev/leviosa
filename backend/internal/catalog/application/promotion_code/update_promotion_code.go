package promotionCode

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
)

func (s *PromotionCodeService) UpdatePromotionCode(ctx context.Context, promotionCodeID string, request *domain.UpdatePromotionCodeRequest) error {
	if err := request.Valid(ctx); err != nil {
		return errs.NewInvalidValueErr(err.Error())
	}

	id, err := uuid.Parse(promotionCodeID)
	if err != nil {
		return errs.NewInvalidValueErr("invalid promotion code ID format")
	}

	// Check if promotion code exists
	_, err = s.repo.GetPromotionCodeByID(ctx, id)
	if err != nil {
		return fmt.Errorf("validate promotion code: %w", err)
	}

	if err := s.repo.UpdatePromotionCode(ctx, id, request); err != nil {
		return fmt.Errorf("update promotion code: %w", err)
	}

	return nil
}
