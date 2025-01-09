package eventRepository

import (
	"context"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain/event"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (e *EventRepository) GetAllEvents(ctx context.Context) ([]*eventService.Event, error) {
	query := `
        SELECT 
            id,
            title,
            description,
            type,
            location,
            placecount,
            begin_at,
            end_at
        FROM events;`
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
		var event eventService.Event
		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.Type,
			&event.Location,
			&event.PlaceCount,
			&event.BeginAtFormatted,
			&event.EndAtFormatted,
		); err != nil {
			return nil, rp.NewDatabaseErr(err)
		}
		event.Parse()
		events = append(events, &event)
	}
	if err = rows.Err(); err != nil {
		return nil, rp.NewDatabaseErr(err)
	}
	return events, nil
}
