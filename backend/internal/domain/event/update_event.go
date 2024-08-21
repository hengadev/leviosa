package event

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) ModifyEvent(ctx context.Context, event *Event) (string, error) {
	err := s.Repo.RemoveEvent(ctx, event.ID)
	if err != nil {
		repository.NewRessourceUpdateErr(err)
	}
	return "", nil
}
