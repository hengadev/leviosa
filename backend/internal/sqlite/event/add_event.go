package eventRepository

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (e *EventRepository) AddEvent(ctx context.Context, event *event.Event) (string, error) {
	fail := func(err error) (string, error) {
		return "", rp.NewRessourceCreationErr(err)
	}
	beginat, err := formatBeginAt(event.BeginAt)
	if err != nil {
		fail(err)
	}
	minutes := int(event.SessionDuration)
	query := "INSERT INTO events (id, location, placecount, freeplace, beginat, sessionduration, priceid, day, month, year) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);"
	res, err := e.DB.ExecContext(ctx, query,
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
		return fail(err)
	}
	lastInsertID, err := res.LastInsertId()
	if err != nil {
		return fail(err)
	}
	if lastInsertID == 0 {
		return fail(errors.New("no insertion in database"))
	}
	return event.ID, nil
}
