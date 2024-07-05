package event

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) ModifyEvent(ctx context.Context, event *Event) (string, error) {
	ID, err := s.Repo.RemoveEvent(ctx, event.ID)
	if err != nil || ID != event.ID {
		repository.NewRessourceUpdateErr(err)
	}
	return "", nil
}
