package eventRepository

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) AddEvent(ctx context.Context, event *eventService.Event) (string, error) {
	query := `INSERT INTO events (
                id,
                title,
                description,
                type,
                location,
                placecount,
                freeplace,
                begin_at,
                end_at,
                session_duration
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := e.DB.ExecContext(ctx, query,
		event.ID,
		event.Title,
		event.Description,
		event.Type,
		event.Location,
		event.PlaceCount,
		event.FreePlace,
		event.BeginAtFormatted,
		event.SessionDuration,
	)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return "", rp.NewContextErr(err)
		default:
			return "", rp.NewDatabaseErr(err)
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", rp.NewDatabaseErr(err)
	}
	if rowsAffected == 0 {
		return "", rp.NewNotCreatedErr(errors.New("no rows affected by insertion statement"), "event")
	}
	return event.ID, nil
}
