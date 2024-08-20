package eventRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) GetEventByID(ctx context.Context, id string) (*event.Event, error) {
	event := &event.Event{}
	var beginat string
	var minutes int
	var err error
	query := "SELECT id, location, placecount, freeplace, beginat, sessionduration, day, month, year FROM events WHERE id=?"
	if err := e.DB.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.Location,
		&event.PlaceCount,
		&event.FreePlace,
		&beginat,
		&minutes,
		&event.Day,
		&event.Month,
		&event.Year,
	); err != nil {
		return nil, rp.NewNotFoundError(err)
	}
	event.SessionDuration = time.Minute * time.Duration(minutes)
	event.BeginAt, err = parseBeginAt(beginat, event.Day, event.Month, event.Year)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "error parsing time", err)
	}
	return event, nil
}
