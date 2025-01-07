package eventService

import (
	"context"
)

type Reader interface {
	GetEventByID(ctx context.Context, eventID string) (*Event, error)
	GetEventForUser(ctx context.Context, userID string) (*EventUser, error)
	GetPriceID(ctx context.Context, eventID string) (string, error)
	EventHasAvailablePlaces(ctx context.Context, eventID string) (bool, error)
}

type Writer interface {
	AddEvent(ctx context.Context, event *Event) error
	RemoveEvent(ctx context.Context, eventID string) error
	ModifyEvent(ctx context.Context, event *Event, whereMap map[string]any, prohibitedFields ...string) error
	DecreaseFreePlace(ctx context.Context, eventID string) error
}
type ReadWriter interface {
	Reader
	Writer
}
