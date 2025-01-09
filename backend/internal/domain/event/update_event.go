package eventService

import (
	"context"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (s *Service) ModifyEvent(ctx context.Context, event *Event) error {
	if event == nil {
		return domain.NewInvalidValueErr("event can not be nil")
	}
	// TODO: get the prohibited fields for the function in here
	whereMap := map[string]any{"id": event.ID}
	if err := s.Repo.ModifyEvent(
		ctx,
		event,
		whereMap,
	); err != nil {
		switch {
		case errors.Is(err, rp.ErrDatabase):
			return domain.NewQueryFailedErr(err)
		case errors.Is(err, rp.ErrInternal):
			return err
		case errors.Is(err, rp.ErrNotUpdated):
			return domain.NewNotUpdatedErr(err)
		case errors.Is(err, rp.ErrContext):
			return err
		default:
			return domain.NewUnexpectTypeErr(err)
		}
	}
	return nil
}
