package eventRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) GetAllEvents(ctx context.Context) ([]*event.Event, error) {
	var events []*event.Event
	query := "SELECT id, location, placecount, freeplace, beginat, sessionduration, day, month, year FROM events;"
	rows, err := e.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, rp.NewErrRow(err)
	}
	defer rows.Close()
	for rows.Next() {
		var beginat string
		var minutes int
		event := &event.Event{}
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
			return nil, rp.NewErrScan(err)
		}
		event.SessionDuration = time.Minute * time.Duration(minutes)
		event.BeginAt, err = parseBeginAt(beginat, event.Day, event.Month, event.Year)
		if err != nil {
			return nil, fmt.Errorf("parsing time: %w", err)
		}
		events = append(events, event)
	}
	return events, nil
}
