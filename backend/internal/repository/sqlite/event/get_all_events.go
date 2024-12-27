package eventRepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) GetAllEvents(ctx context.Context) ([]*eventService.Event, error) {
	query := "SELECT id, location, placecount, freeplace, beginat, sessionduration, day, month, year FROM events;"
	rows, err := e.DB.QueryContext(ctx, query)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextErr(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	defer rows.Close()
	var events []*eventService.Event
	for rows.Next() {
		var beginat string
		var minutes int
		event := &eventService.Event{}
		if err := rows.Scan(
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
			return nil, rp.NewDatabaseErr(err)
		}
		event.SessionDuration = time.Minute * time.Duration(minutes)
		event.BeginAt, err = parseBeginAt(beginat, event.Day, event.Month, event.Year)
		if err != nil {
			return nil, rp.NewInternalErr(fmt.Errorf("parsing time: %w", err))
		}
		events = append(events, event)
	}
	if err = rows.Err(); err != nil {
		return nil, rp.NewDatabaseErr(err)
	}
	return events, nil
}
