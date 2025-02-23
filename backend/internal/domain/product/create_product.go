package productService

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/leviosa/internal/domain"
	rp "github.com/GaryHY/leviosa/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) CreateProduct(ctx context.Context, product *Product) error {
	if errs := product.Valid(ctx); len(errs) > 0 {
		return domain.NewInvalidValueErr(fmt.Sprintf("product validation error: %s", errs.Error()))
	}
	product.ID = uuid.NewString()
	if err := s.repo.AddProduct(ctx, product); err != nil {
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
