package productService

import (
	"context"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (s *Service) RemoveProduct(ctx context.Context, productID string) error {
	err := s.repo.RemoveProduct(ctx, productID)
	if err != nil {
		// TODO: do all the error handling for that thing
		switch {
		case errors.Is(err, rp.ErrDatabase):
			return domain.NewQueryFailedErr(err)
		case errors.Is(err, rp.ErrContext):
			return err
		default:
			return domain.NewUnexpectTypeErr(err)
		}
	}
	return nil
}
