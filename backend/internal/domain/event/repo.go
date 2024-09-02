package eventService

import (
	"context"
)

type Reader interface {
	GetEventByID(ctx context.Context, eventID string) (*Event, error)
	GetEventForUser(ctx context.Context, userID int) (*EventUser, error)
	GetPriceIDByEventID(ctx context.Context, eventID string) (string, error)
}

type Writer interface {
	AddEvent(ctx context.Context, event *Event) error
	RemoveEvent(ctx context.Context, eventID string) error
	ModifyEvent(ctx context.Context, event *Event, whereMap map[string]any, prohibitedFields ...string) error
	DecreaseFreeplace(ctx context.Context, eventID string) error
}
type ReadWriter interface {
	Reader
	Writer
}
