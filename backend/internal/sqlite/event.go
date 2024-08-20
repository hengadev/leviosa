package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

type EventRepository struct {
	DB *sql.DB
}

func (e *EventRepository) GetDB() *sql.DB {
	return e.DB
}

func NewEventRepository(ctx context.Context, db *sql.DB) *EventRepository {
	return &EventRepository{db}
}

// the new functions
func (e *EventRepository) GetEventByID(ctx context.Context, id string) (*event.Event, error) {
	event := &event.Event{}
	var beginat string
	var minutes int
	var err error
	query := "SELECT id, location, placecount, freeplace, beginat, sessionduration, day, month, year FROM events WHERE id=?"
	if err := e.DB.QueryRowContext(ctx, query, id).Scan(
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
		return nil, rp.NewNotFoundError(err)
	}
	event.SessionDuration = time.Minute * time.Duration(minutes)
	event.BeginAt, err = parseBeginAt(beginat, event.Day, event.Month, event.Year)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "error parsing time", err)
	}
	return event, nil
}

func (e *EventRepository) GetEventByUserID(ctx context.Context, userID string) ([]*event.Event, error) {
	events := make([]*event.Event, 0)
	query := `
       SELECT * FROM events WHERE id IN
       (SELECT eventid FROM votes WHERE userid=?)
       ORDER BY rowid ASC;
	   `
	rows, err := e.DB.QueryContext(ctx, query, userID)
	defer rows.Close()
	if err != nil {
		return nil, rp.NewErrRow(err)
	}

	for rows.Next() {
		event := &event.Event{}
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

func (e *EventRepository) DecreaseFreeplace(ctx context.Context, ID string) (int, error) {
	res, err := e.DB.ExecContext(ctx, "UPDATE events SET freeplace = freeplace - 1 WHERE id=?;", ID)
	if err != nil {
		return 0, rp.NewRessourceUpdateErr(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, rp.NewRessourceUpdateErr(err)
	}
	if rowsAffected == 0 {
		return 0, fmt.Errorf("rowsAffected = 0, ID not found")
	}
	return int(rowsAffected), nil
}

// old functions
func (e *EventRepository) AddEvent(ctx context.Context, event *event.Event) (string, error) {
	new_date := event.BeginAt.Format(time.RFC3339)
	_, err := e.DB.ExecContext(
		ctx,
		"INSERT INTO events (id, location, placecount, date, priceid) VALUES (?, ?, ?, ?, ?)",
		event.ID,
		event.Location,
		event.PlaceCount,
		new_date,
		event.PriceID,
	)
	if err != nil {
		return "", rp.NewRessourceCreationErr(err)
	}
	return event.ID, nil
}

func (e *EventRepository) ModifyEvent(ctx context.Context, event *event.Event) (*event.Event, error) {
	_, err := e.DB.ExecContext(
		ctx,
		"UPDATE events SET location=?, placecount=?, date=? WHERE id=?;",
		event.Location,
		event.PlaceCount,
		event.BeginAt.Format(time.RFC3339),
		event.ID,
	)
	if err != nil {
		return nil, rp.NewRessourceUpdateErr(err)
	}
	return event, nil
}

func (e *EventRepository) RemoveEvent(ctx context.Context, eventID string) (string, error) {
	_, err := e.DB.ExecContext(ctx, "DELETE from events where id=?;", eventID)
	if err != nil {
		return "", rp.NewRessourceDeleteErr(err)
	}
	return eventID, nil
}

// Function that returns true if an event with the ID "eventID" is in the database and if the number of place found in "placecount" is > 0.
func (e *EventRepository) CheckEvent(ctx context.Context, eventID string) (bool, error) {
	var placecount int
	err := e.DB.QueryRowContext(ctx, "SELECT placecount FROM events WHERE id=?;", eventID).Scan(&placecount)
	if err == sql.ErrNoRows {
		return false, rp.NewNotFoundError(err)
	}
	if err != nil {
		return false, rp.NewBadQueryErr(err)
	}
	return placecount > 0, nil
}

func (e *EventRepository) DecreaseEventPlacecount(ctx context.Context, eventID string) error {
	_, err := e.DB.ExecContext(ctx, "UPDATE events SET placecount = placecount-1 WHERE id=?", eventID)
	if err != nil {
		return rp.NewRessourceUpdateErr(err)
	}
	return nil
}

func (e *EventRepository) GetPriceIDByEventID(ctx context.Context, eventID string) (string, error) {
	var priceID string
	err := e.DB.QueryRowContext(ctx, "SELECT priceid from events where id = ?;", eventID).Scan(&priceID)
	if err == sql.ErrNoRows {
		return "", rp.NewNotFoundError(err)
	}
	if err != nil {
		return "", rp.NewBadQueryErr(err)
	}
	return priceID, nil
}

// On part du principe que le beginAt est store comme "xx:xx:xx"

func (e *EventRepository) GetEventForUser(ctx context.Context, userID string) (*event.EventUser, error) {
	// TODO: use transaction for that function brother
	var res event.EventUser
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
			event := &event.Event{}
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

func convIntToStr(value int) (string, error) {
	if value < 0 {
		return "", fmt.Errorf("%d is invalid value", value)
	} else if value < 10 {
		return fmt.Sprintf("0%d", value), nil
	} else if value < 100 {
		return fmt.Sprintf("%d", value), nil
	} else {
		return "", fmt.Errorf("%d is invalid value", value)
	}
}

// helper function for the GetEventForUser function
func formatTime(hour string) (string, error) {
	res := hour
	suffix := "AM"
	timeHour, err := time.Parse(time.TimeOnly, hour)
	if err != nil {
		return "", fmt.Errorf("error parsing time: %w", err)
	}
	if timeHour.Hour() > 12 {
		suffix = "PM"
		hour, err := convIntToStr(timeHour.Hour() - 12)
		if err != nil {
			return "", fmt.Errorf("convert string to int: %w", err)
		}
		minute, err := convIntToStr(timeHour.Minute())
		if err != nil {
			return "", fmt.Errorf("convert string to int: %w", err)
		}
		second, err := convIntToStr(timeHour.Second())
		if err != nil {
			return "", fmt.Errorf("convert string to int: %w", err)
		}

		res = fmt.Sprintf("%s:%s:%s", hour, minute, second)
	}
	return res + suffix, nil
}

// helper function for the GetEventForUser function
func parseBeginAt(hour string, day, month, year int) (time.Time, error) {
	var res time.Time
	hourFormatted, err := formatTime(hour)
	if err != nil {
		return res, err
	}
	parsedDay, err := convIntToStr(day)
	if err != nil {
		return res, fmt.Errorf("convert string to int: %w", err)
	}
	parsedMonth, err := convIntToStr(month)
	if err != nil {
		return res, fmt.Errorf("convert string to int: %w", err)
	}
	parsedYear, err := convIntToStr(year % 100)
	if err != nil {
		return res, fmt.Errorf("convert string to int: %w", err)
	}
	dateFormatted := fmt.Sprintf("%s/%s %s '%s -0700", parsedMonth, parsedDay, hourFormatted, parsedYear)
	res, err = time.Parse(time.Layout, dateFormatted)
	if err != nil {
		return res, fmt.Errorf("parsing formatted date: %w", err)
	}
	return res, nil
}
