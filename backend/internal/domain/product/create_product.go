package productService

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) CreateProduct(ctx context.Context, product *Product) error {
	if errs := product.Valid(ctx); len(errs) > 0 {
		return domain.NewInvalidValueErr(fmt.Sprintf("product validation error: %s", errs.Error()))
	}
	if err := s.repo.AddProduct(ctx, product); err != nil {
		switch {
		// TODO: add other errors
		case errors.Is(err, rp.ErrNotCreated):
			return domain.NewNotCreatedErr(err)
		}
	}
	return nil
}
