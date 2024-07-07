package event

import (
	"context"
)

type Reader interface {
	GetEventByID(ctx context.Context, eventID string) (*Event, error)
	GetEventForUser(ctx context.Context, userID string) (*EventUser, error)
	GetPriceIDByEventID(ctx context.Context, eventID string) (string, error)
	// GetTimeInfoByID(ctx context.Context, eventID string) (time.Time, time.Duration, int, int, int, error)
}

type Writer interface {
	AddEvent(ctx context.Context, event *Event) (string, error)
	RemoveEvent(ctx context.Context, eventID string) (string, error)
	ModifyEvent(ctx context.Context, eventID string) (*Event, error)
	DecreaseFreeplace(ctx context.Context, eventID string) error
}
type ReadWriter interface {
	Reader
	Writer
}
