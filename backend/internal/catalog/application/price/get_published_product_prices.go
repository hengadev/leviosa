package price

import (
	"context"
	"errors"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
)

// GetPublishedProductPrices returns the active prices of a published product,
// for the public price route. A draft or archived product is not found, the
// same answer as for a missing one.
func (s *PriceService) GetPublishedProductPrices(ctx context.Context, productIDStr string) ([]*domain.Price, error) {
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return nil, errs.NewInvalidValueErr(fmt.Sprintf("product ID is invalid: %s", err.Error()))
	}

	product, err := s.sharedRepo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to check product existence: %w", err)
	}
	if product.Status != domain.Published {
		return nil, errs.NewNotFoundErr(errors.New("product is not published"), "product")
	}

	prices, err := s.repo.GetPricesByProductID(ctx, productIDStr, true)
	if err != nil {
		return nil, fmt.Errorf("failed to list prices: %w", err)
	}

	return prices, nil
}
