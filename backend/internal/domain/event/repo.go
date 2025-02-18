package eventService

import (
	"context"

	"github.com/GaryHY/leviosa/internal/domain/event/models"
)

type Reader interface {
	GetEventByID(ctx context.Context, eventID string) (*models.Event, error)
	GetEventForUser(ctx context.Context, userID string) (*models.EventUser, error)
	GetPriceID(ctx context.Context, eventID string) (string, error)
	EventHasAvailablePlaces(ctx context.Context, eventID string) (bool, error)
}

type Writer interface {
	AddEvent(ctx context.Context, event *models.Event) (string, error)
	RemoveEvent(ctx context.Context, eventID string) error
	ModifyEvent(ctx context.Context, event *models.Event, whereMap map[string]any) error
	DecreaseFreePlace(ctx context.Context, eventID string) error
}
type ReadWriter interface {
	Reader
	Writer
}
