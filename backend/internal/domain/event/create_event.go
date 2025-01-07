package eventService

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) CreateEvent(ctx context.Context, event *Event) (string, error) {
	// TODO:
	// - check if date is available in the database
	if err := event.Format(ctx); err != nil {
		return "", domain.NewFormatError("event", err)
	}
	eventID, err := s.Repo.AddEvent(ctx, event)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrNotCreated):
			return "", domain.NewNotCreatedErr(err)
		case errors.Is(err, rp.ErrDatabase):
			return "", domain.NewQueryFailedErr(err)
		case errors.Is(err, rp.ErrContext):
			return "", err
		default:
			return "", domain.NewUnexpectTypeErr(err)
		}
	}
	return eventID, nil
}
