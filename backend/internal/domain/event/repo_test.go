package eventService_test

import (
	"context"
	// "fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
)

type StubEventRepository struct {
	events []*eventService.Event
}

func NewStubEventRepository() *StubEventRepository {
	return &StubEventRepository{}
}

func (s *StubEventRepository) GetEventByID(ctx context.Context, eventID string) (*eventService.Event, error) {
	for _, event := range s.events {
		if event.ID == eventID {
			return event, nil
		}
	}
	return nil, nil
}

func (s *StubEventRepository) GetEventForUser(ctx context.Context, userID string) (*eventService.EventUser, error) {
	return nil, nil
}

func (s *StubEventRepository) GetPriceIDByEventID(ctx context.Context, eventID string) (string, error) {

	for _, event := range s.events {
		if event.ID == eventID {
			return event.PriceID, nil
		}
	}
	return "", nil
}

func (s *StubEventRepository) AddEvent(ctx context.Context, event *eventService.Event) error {
	s.events = append(s.events, event)
	return nil
}
func (s *StubEventRepository) RemoveEvent(ctx context.Context, eventID string) error {
	return nil
}
func (s *StubEventRepository) ModifyEvent(ctx context.Context, event *eventService.Event, whereMap map[string]any, prohibitedFields ...string) error {
	return nil
}
func (s *StubEventRepository) DecreaseFreeplace(ctx context.Context, eventID string) error {
	return nil
}
