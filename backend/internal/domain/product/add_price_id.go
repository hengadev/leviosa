package productService

import (
	"context"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (s *Service) AddPriceID(ctx context.Context, productID, priceID string) error {
	if err := s.repo.AddPriceID(ctx, productID, priceID); err != nil {
		switch {
		case errors.Is(err, rp.ErrContext):
			return err
		case errors.Is(err, rp.ErrDatabase):
			return domain.NewQueryFailedErr(err)
		case errors.Is(err, rp.ErrNotUpdated):
			return domain.NewNotUpdatedErr(err)
		default:
			return domain.NewUnexpectTypeErr(err)
		}
	}
	return nil
}
