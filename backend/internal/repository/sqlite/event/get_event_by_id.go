package eventRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain/event"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (e *EventRepository) GetEventByID(ctx context.Context, id string) (*eventService.Event, error) {
	event := &eventService.Event{}
	query := `
        SELECT 
            title,
            description,
            type,
            location,
            placecount,
            freeplace,
            begin_at,
            end_at
        FROM events 
        WHERE id = ?;`

	if err := e.DB.QueryRowContext(ctx, query, id).Scan(
		&event.Title,
		&event.Description,
		&event.Type,
		&event.Location,
		&event.PlaceCount,
		&event.FreePlace,
		&event.BeginAtFormatted,
		&event.EndAtFormatted,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, rp.NewNotFoundErr(err, "user")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextErr(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	// parse event to make it application ready
	event.Parse()
	return event, nil
}
