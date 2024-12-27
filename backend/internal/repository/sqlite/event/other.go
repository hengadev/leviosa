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

// NOTE: just a a collection of function that might be useful, I do not want to code them again

// Function that returns true if an event with the ID "eventID" is in the database and if the number of place found in "placecount" is > 0.
func (e *EventRepository) CheckEvent(ctx context.Context, eventID string) (bool, error) {
	var placecount int
	err := e.DB.QueryRowContext(ctx, "SELECT placecount FROM events WHERE id=?;", eventID).Scan(&placecount)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, rp.NewNotFoundErr(err, "event")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return false, rp.NewContextErr(err)
		default:
			return false, rp.NewDatabaseErr(err)
		}
	}
	return placecount > 0, nil
}

func (e *EventRepository) DecreaseEventPlacecount(ctx context.Context, eventID string) error {
	result, err := e.DB.ExecContext(ctx, "UPDATE events SET placecount = placecount-1 WHERE id=?", eventID)
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
		return rp.NewNotFoundErr(fmt.Errorf("no rows affected"), "event")
	}
	return nil
}

// On part du principe que le beginAt est store comme "xx:xx:xx"

// NOTE: votes do not include event id now
func (e *EventRepository) GetEventByUserID(ctx context.Context, userID string) ([]*eventService.Event, error) {
	query := `
       SELECT * FROM events WHERE id IN
       (SELECT eventid FROM votes WHERE userid=?)
       ORDER BY rowid ASC;
	   `
	rows, err := e.DB.QueryContext(ctx, query, userID)
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
		event := &eventService.Event{}
		var dataTemp string
		if err := rows.Scan(&event.ID, &event.Location, &event.PlaceCount, &dataTemp, &event.PriceID); err != nil {
			return nil, rp.NewNotFoundErr(err, "event")
		}
		event.BeginAt, err = time.Parse(time.RFC3339, dataTemp)
		if err != nil {
			return nil, rp.NewInternalErr(fmt.Errorf("%s: %w", "error parsing time", err))
		}
		events = append(events, event)
	}
	return events, nil
}
