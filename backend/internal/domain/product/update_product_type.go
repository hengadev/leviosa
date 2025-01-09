package productService

import (
	"context"

	"github.com/GaryHY/leviosa/internal/domain"
)

func (s *Service) UpdateProductType(ctx context.Context, product *ProductType) error {
	if err := s.repo.ModifyProductType(
		ctx,
		product,
		map[string]any{"id": product.ID},
		"id",
	); err != nil {
		switch {
		default:
			return domain.NewUnexpectTypeErr(err)
		}
	}
	return nil
}
