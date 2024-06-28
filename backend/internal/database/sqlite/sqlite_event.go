package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/types"
	// "github.com/aws/aws-sdk-go-v2/aws/protocol/query"
)

func (s *Store) GetEventByID(id string) *types.Event {
	event := &types.Event{}
	if err := s.DB.QueryRow("SELECT * FROM events WHERE id=?;", id).Scan(&event.Id, &event.Location, &event.PlaceCount, &event.BeginAt); err != nil {
		return event
	}
	return event
}

// func (s *Store) GetEventByUserId(ctx context.Context, user_id string) []*types.Event {
func (s *Store) GetEventByUserId(userID string) []*types.Event {
	events := make([]*types.Event, 0)
	query := `
       SELECT * FROM events WHERE id IN
       (SELECT eventid FROM votes WHERE userid=?)
       ORDER BY rowid ASC;
	   `
	rows, err := s.DB.Query(query, userID)
	defer rows.Close()
	if err != nil {
		log.Fatal("Cannot get events rows - ", err)
	}

	for rows.Next() {
		event := &types.Event{}
		var dataTemp string
		if err := rows.Scan(&event.Id, &event.Location, &event.PlaceCount, &dataTemp, &event.PriceId); err != nil {
			log.Fatal("Cannot scan the event - ", err)
		}
		event.BeginAt, err = time.Parse(time.RFC3339, dataTemp)
		if err != nil {
			log.Fatal("Failed to parse the date from the database in the GetUserByEventId function - ", err)
		}
		events = append(events, event)
	}
	return events
}

func (s *Store) GetAllEvents() []*types.Event {
	events := make([]*types.Event, 0)
	rows, err := s.DB.Query("SELECT * FROM events;")
	if err != nil {
		log.Fatal("Cannot get events rows - ", err)
	}
	defer rows.Close()
	var dateTemp string

	for rows.Next() {
		event := &types.Event{}
		if err := rows.Scan(&event.Id, &event.Location, &event.PlaceCount, &dateTemp, &event.PriceId); err != nil {
			log.Fatal("Cannot scan the event - ", err)
		}
		// NOTE: Old version but now I am using another formatting to send to the frontend
		// event.BeginAt, err = time.Parse(types.EventFormat, dateTemp)
		event.BeginAt, err = time.Parse(time.RFC3339, dateTemp)
		if err != nil {
			log.Fatal("Failed to parse the date from the database - ", err)
		}
		events = append(events, event)
	}
	return events
}

func (s *Store) PostEvent(event *types.Event) {
	new_date := event.BeginAt.Format(time.RFC3339)
	_, err := s.DB.Exec("INSERT INTO events (id, location, placecount, date, priceid) VALUES (?, ?, ?, ?, ?)", event.Id, event.Location, event.PlaceCount, new_date, event.PriceId)
	// _, err := s.DB.Exec("INSERT INTO events (id, location, placecount, date, priceid) VALUES (?, ?, ?, ?, ?)", event.Id, event.Location, event.PlaceCount, event.BeginAt, event.PriceId)
	if err != nil {
		log.Fatal("Could not insert new event into the database - ", err)
	}
}

func (s *Store) UpdateEvent(event *types.Event) error {
	_, err := s.DB.Exec("UPDATE events SET location=?, placecount=?, date=? WHERE id=?;", event.Location, event.PlaceCount, event.BeginAt.Format(time.RFC3339), event.Id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteEvent(eventId string) error {
	_, err := s.DB.Exec("DELETE from events where id=?;", eventId)
	if err != nil {
		return err
	}
	return nil
}

// Function that returns true if an event with the ID "event_id" is in the database and if the number of place found in "placecount" is > 0.
func (s *Store) CheckEvent(eventId string) bool {
	var placecount int
	err := s.DB.QueryRow("SELECT placecount FROM events WHERE id=?;", eventId).Scan(&placecount)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		log.Fatalf("Could not select to see if event with id %q exists - %s", eventId, err)
	}
	return true && placecount > 0
}

func (s *Store) DecreaseEventPlacecount(eventId string) error {
	_, err := s.DB.Exec("UPDATE events SET placecount = placecount-1 WHERE id=?", eventId)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetPriceIDByEventID(eventId string) (priceId string) {
	s.DB.QueryRow("SELECT priceid from events where id = ?;", eventId).Scan(&priceId)
	return
}

// On part du principe que le beginAt est store comme "xx:xx:xx"

func (s *Store) GetEventForUser(userId string) (types.EventBody, error) {
	var res types.EventBody
	now := time.Now()
	day, month, year := now.Day(), int(now.Month()), now.Year()
	statements := []struct {
		condition string
		field     string
	}{
		{condition: fmt.Sprintf("(month < %d AND year = %d) OR (year < %d) OR (day < %d AND month = %d AND year = %d) LIMIT 3", month, year, year, day, month, year), field: "past"},
		{condition: fmt.Sprintf("(year > %d) OR (month > %d AND year = %d) OR (day > %d AND month = %d AND year = %d) LIMIT 3", year, month, year, day, month, year), field: "next"},
		// TODO: make that condition right so that I get the right incoming events
		// {condition: fmt.Sprintf("(month > %d AND year = %d) OR (day > %d AND month = %d AND year = %d) LIMIT=1", month, year, day, month, year), field: "incoming"},
		{condition: fmt.Sprintf("(day > %d AND month = %d AND year = %d) OR (month = %d + 1 AND year = %d) OR (month = 1 AND year = %d + 1) LIMIT 1", day, month, year, year, month, year), field: "incoming"},
	}

	// PERF: Can I access this concurrently for speed purposes ?
	for _, statement := range statements {
		query := fmt.Sprintf("SELECT * FROM events WHERE %s;", statement.condition)
		rows, err := s.DB.Query(query, userId)
		defer rows.Close()
		if err != nil {
			return res, err
		}
		for rows.Next() {
			var priceId string
			var beginAt string
			event := &types.EventSent{}
			if err := rows.Scan(
				&event.Id,
				&event.Location,
				&event.PlaceCount,
				&beginAt,
				&event.SessionDuration,
				&priceId,
				&event.Day,
				&event.Month,
				&event.Year,
			); err != nil {
				return res, fmt.Errorf("Failed to scan events: %w", err)
			}
			event.BeginAt, err = parseBeginAt(beginAt, event.Day, event.Month, event.Year)
			if err != nil {
				return res, fmt.Errorf("Failed to parse the beginAt field : %w", err)
			}
			var usedCount int
			query := fmt.Sprintf("SELECT COUNT(userid) from event_%s;", event.Id)
			if err := s.DB.QueryRow(query).Scan(&usedCount); err != nil {
				return res, fmt.Errorf("Failed to get the FreePlace field for the event: %w", err)
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
	return res, nil
}

// TODO:
// admin stores the BeginAt as some time in a specific format
// I get here the time that I parse since it is going to be written as what ?

// helper function for the GetEventForUser function
func convertToString(value int) string {
	var res string
	if value < 10 {
		res = fmt.Sprintf("0%d", value)
	} else {
		res = fmt.Sprintf("%d", value)
	}
	return res
}

// helper function for the GetEventForUser function
func convertToStringHour(hour string) (string, error) {
	res := hour
	suffix := "AM"
	timeHour, err := time.Parse(time.TimeOnly, hour)
	if err != nil {
		return "", fmt.Errorf("got some error mate : %w", err)
	}
	if timeHour.Hour() > 12 {
		suffix = "PM"
		res = fmt.Sprintf("%s:%s:%s", convertToString(timeHour.Hour()-12), convertToString(timeHour.Minute()), convertToString(timeHour.Second()))
	}
	return res + suffix, nil
}

// helper function for the GetEventForUser function
func parseBeginAt(hour string, day, month, year int) (time.Time, error) {
	var res time.Time
	hourFormatted, err := convertToStringHour(hour)
	if err != nil {
		return res, err
	}
	dateFormatted := fmt.Sprintf("%s/%s %s '%s -0700", convertToString(month), convertToString(day), hourFormatted, convertToString(year%100))
	res, err = time.Parse(time.Layout, dateFormatted)
	if err != nil {
		fmt.Println("got some error mate : ", err)
	}
	return res, nil
}

// # table eventid
// - userid references users
// - creneau (quand est ce que tu commences ?)
