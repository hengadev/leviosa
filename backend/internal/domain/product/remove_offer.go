package productService

import (
	"context"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (s *Service) RemoveOffer(ctx context.Context, offerID int) error {
	if err := s.repo.RemoveOffer(ctx, offerID); err != nil {
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
