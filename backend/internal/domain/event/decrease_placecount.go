package eventService

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) DecreasePlacecount(ctx context.Context, eventID string) error {
	err := s.Repo.DecreaseFreeplace(ctx, eventID)
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
