package eventRepository

import (
	"context"
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
		return false, rp.NewQueryErr(err)
	}
	return placecount > 0, nil
}

func (e *EventRepository) DecreaseEventPlacecount(ctx context.Context, eventID string) error {
	res, err := e.DB.ExecContext(ctx, "UPDATE events SET placecount = placecount-1 WHERE id=?", eventID)
	if err != nil {
		return rp.NewRessourceUpdateErr(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return rp.NewRessourceUpdateErr(err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no row updated")
	}
	return nil
}

// On part du principe que le beginAt est store comme "xx:xx:xx"

// NOTE: votes do not include event id now
func (e *EventRepository) GetEventByUserID(ctx context.Context, userID string) ([]*eventService.Event, error) {
	events := make([]*eventService.Event, 0)
	query := `
       SELECT * FROM events WHERE id IN
       (SELECT eventid FROM votes WHERE userid=?)
       ORDER BY rowid ASC;
	   `
	rows, err := e.DB.QueryContext(ctx, query, userID)
	defer rows.Close()
	if err != nil {
		return nil, rp.NewDatabaseErr(err)
	}

	for rows.Next() {
		event := &eventService.Event{}
		var dataTemp string
		if err := rows.Scan(&event.ID, &event.Location, &event.PlaceCount, &dataTemp, &event.PriceID); err != nil {
			return nil, rp.NewErrScan(err)
		}
		event.BeginAt, err = time.Parse(time.RFC3339, dataTemp)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", "error parsing time", err)
		}
		events = append(events, event)
	}
	return events, nil
}
