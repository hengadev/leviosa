package eventRepository

import (
	"context"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain/event/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (e *EventRepository) GetAllEvents(ctx context.Context) ([]*models.Event, error) {
	query := `
        SELECT 
            id,
            title,
            description,
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
	var events []*models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.PlaceCount,
			&event.EncryptedBeginAt,
			&event.EncryptedEndAt,
		); err != nil {
			return nil, rp.NewDatabaseErr(err)
		}
		events = append(events, &event)
	}
	if err = rows.Err(); err != nil {
		return nil, rp.NewDatabaseErr(err)
	}
	return events, nil
}
