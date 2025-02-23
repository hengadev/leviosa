package productService

import (
	"context"

	"github.com/GaryHY/leviosa/internal/domain"
)

func (s *Service) UpdateProductType(ctx context.Context, product *Offer) error {
	if err := s.repo.ModifyOffer(
		ctx,
		product,
		map[string]any{"id": product.ID},
	); err != nil {
		switch {
		default:
			return domain.NewUnexpectTypeErr(err)
		}
	}
	return nil
}
