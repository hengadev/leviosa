package eventRepository

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) AddEvent(ctx context.Context, event *eventService.Event) error {
	beginat, err := formatBeginAt(event.BeginAt)
	if err != nil {
		rp.NewInternalErr(err)
	}
	minutes := int(event.SessionDuration)
	query := "INSERT INTO events (id, location, placecount, freeplace, beginat, sessionduration, priceid, day, month, year) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);"
	result, err := e.DB.ExecContext(ctx, query,
		event.ID,
		event.Location,
		event.PlaceCount,
		event.FreePlace,
		beginat,
		minutes,
		event.PriceID,
		event.Day,
		event.Month,
		event.Year,
	)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	if rowsAffected == 0 {
		return rp.NewNotCreatedErr(errors.New("no rows affected by insertion statement"), "event")
	}
	return nil
}
