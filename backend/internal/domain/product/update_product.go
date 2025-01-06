package productService

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain"
)

func (s *Service) UpdateProduct(ctx context.Context, product *Product) error {
	if err := s.repo.ModifyProduct(
		ctx,
		product,
		map[string]any{"id": product.ID},
		"id",
		"priceID",
	); err != nil {
		switch {
		default:
			return domain.NewUnexpectTypeErr(err)
		}
	}
	return nil
}
