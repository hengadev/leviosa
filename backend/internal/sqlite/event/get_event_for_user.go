package eventRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) GetEventForUser(ctx context.Context, userID int) (*eventService.EventUser, error) {
	// TODO: use transaction for that function brother
	var res eventService.EventUser
	now := time.Now()
	day, month, year := now.Day(), int(now.Month()), now.Year()
	statements := []struct {
		condition string
		field     string
	}{
		{condition: fmt.Sprintf("(month < %d AND year = %d) OR (year < %d) OR (day < %d AND month = %d AND year = %d) LIMIT 3", month, year, year, day, month, year), field: "past"},
		{condition: fmt.Sprintf("(year > %d) OR (month > %d AND year = %d) OR (day > %d AND month = %d AND year = %d) LIMIT 3", year, month, year, day, month, year), field: "next"},
		{condition: fmt.Sprintf("(day > %d AND month = %d AND year = %d) OR (month = %d + 1 AND year = %d) OR (month = 1 AND year = %d + 1) LIMIT 1", day, month, year, year, month, year), field: "incoming"},
	}

	for _, statement := range statements {
		query := fmt.Sprintf("SELECT * FROM events WHERE %s;", statement.condition)
		rows, err := e.DB.QueryContext(ctx, query, userID)
		defer rows.Close()
		if err != nil {
			return &res, err
		}
		for rows.Next() {
			var priceID string
			var beginAt string
			event := &eventService.Event{}
			if err := rows.Scan(
				&event.ID,
				&event.Location,
				&event.PlaceCount,
				&beginAt,
				&event.SessionDuration,
				&priceID,
				&event.Day,
				&event.Month,
				&event.Year,
			); err != nil {
				return &res, rp.NewErrScan(err)
			}
			event.BeginAt, err = parseBeginAt(beginAt, event.Day, event.Month, event.Year)
			if err != nil {
				return &res, fmt.Errorf("%s: %w", "error parsing time", err)
			}
			var usedCount int
			query := fmt.Sprintf("SELECT COUNT(userid) from event_%s;", event.ID)
			if err := e.DB.QueryRowContext(ctx, query).Scan(&usedCount); err != nil {
				return &res, rp.NewNotFoundError(err)
			}
			event.FreePlace = event.PlaceCount - usedCount
			switch statement.field {
			case "past":
				res.PastEvents = append(res.PastEvents, event)
			case "next":
				res.NextEvents = append(res.NextEvents, event)
			case "incoming":
				res.IncomingEvents = append(res.IncomingEvents, event)
			}
		}
	}
	return &res, nil
}
