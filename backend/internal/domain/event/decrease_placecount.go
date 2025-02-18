package eventService

import (
	"context"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (s *Service) DecreasePlacecount(ctx context.Context, eventID string) error {
	err := s.repo.DecreaseFreePlace(ctx, eventID)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrContext):
			return err
		case errors.Is(err, rp.ErrDatabase):
			return domain.NewQueryFailedErr(err)
		case errors.Is(err, rp.ErrNotUpdated):
			domain.NewNotUpdatedErr(err)
		default:
			return domain.NewUnexpectTypeErr(err)
		}
	}
	return nil
}
