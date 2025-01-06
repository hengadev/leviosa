package productService

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) CreateProductType(ctx context.Context, productType *ProductType) error {
	if errs := productType.Valid(ctx); len(errs) > 0 {
		return domain.NewInvalidValueErr(fmt.Sprintf("product type validation error: %s", errs.Error()))
	}
	if err := s.repo.AddProductType(ctx, productType); err != nil {
		switch {
		case errors.Is(err, rp.ErrNotCreated):
			return domain.NewNotCreatedErr(err)
		case errors.Is(err, rp.ErrDatabase):
			return domain.NewQueryFailedErr(err)
		case errors.Is(err, rp.ErrContext):
			return err
		}
	}
	return nil
}
