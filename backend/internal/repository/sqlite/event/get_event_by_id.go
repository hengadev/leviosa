package eventRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) GetEventByID(ctx context.Context, id string) (*eventService.Event, error) {
	event := &eventService.Event{}
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
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, rp.NewNotFoundError(err, "user")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextError(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	event.SessionDuration = time.Minute * time.Duration(minutes)
	event.BeginAt, err = parseBeginAt(beginat, event.Day, event.Month, event.Year)
	if err != nil {
		return nil, rp.NewInternalError(fmt.Errorf("%s: %w", "error parsing time", err))
	}
	return event, nil
}
