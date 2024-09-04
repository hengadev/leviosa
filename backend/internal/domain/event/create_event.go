package eventService

import (
	"context"
	"fmt"
)

func (s *Service) CreateEvent(ctx context.Context, event *Event) (string, error) {
	// TODO:
	// - check if the date is available
	// - create the event using the right function from the repo
	if err := s.Repo.AddEvent(ctx, event); err != nil {
		return "", fmt.Errorf("create event: %s", err)
	}
	// - return the event id and no error
	return event.ID, nil
}
