package sqlite

import (
	"database/sql"
	"log"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/types"
)

func (s *Store) GetEventByID(id string) *types.Event {
	event := &types.Event{}
	if err := s.DB.QueryRow("SELECT * FROM events WHERE id=?;", id).Scan(&event.Id, &event.Location, &event.PlaceCount, &event.Date); err != nil {
		return event
	}
	return event
}

func (s *Store) GetEventByUserId(user_id string) []*types.Event {
	events := make([]*types.Event, 0)
	query := `
       SELECT * FROM events WHERE id IN
       (SELECT eventid FROM votes WHERE userid=?)
       ORDER BY rowid ASC;
	   `
	rows, err := s.DB.Query(query, user_id)
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
		event.Date, err = time.Parse(time.RFC3339, dataTemp)
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
		// event.Date, err = time.Parse(types.EventFormat, dateTemp)
		event.Date, err = time.Parse(time.RFC3339, dateTemp)
		if err != nil {
			log.Fatal("Failed to parse the date from the database - ", err)
		}
		events = append(events, event)
	}
	return events
}

func (s *Store) PostEvent(event *types.Event) {
	new_date := event.Date.Format(time.RFC3339)
	_, err := s.DB.Exec("INSERT INTO events (id, location, placecount, date, priceid) VALUES (?, ?, ?, ?, ?)", event.Id, event.Location, event.PlaceCount, new_date, event.PriceId)
	// _, err := s.DB.Exec("INSERT INTO events (id, location, placecount, date, priceid) VALUES (?, ?, ?, ?, ?)", event.Id, event.Location, event.PlaceCount, event.Date, event.PriceId)
	if err != nil {
		log.Fatal("Could not insert new event into the database - ", err)
	}
}

func (s *Store) UpdateEvent(event *types.Event) error {
	_, err := s.DB.Exec("UPDATE events SET location=?, placecount=?, date=? WHERE id=?;", event.Location, event.PlaceCount, event.Date.Format(time.RFC3339), event.Id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteEvent(event_id string) error {
	_, err := s.DB.Exec("DELETE from events where id=?;", event_id)
	if err != nil {
		return err
	}
	return nil
}

// Function that returns true if an event with the ID "event_id" is in the database and if the number of place found in "placecount" is > 0.
func (s *Store) CheckEvent(event_id string) bool {
	var placecount int
	err := s.DB.QueryRow("SELECT placecount FROM events WHERE id=?;", event_id).Scan(&placecount)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		log.Fatalf("Could not select to see if event with id %q exists - %s", event_id, err)
	}
	return true && placecount > 0
}

func (s *Store) DecreaseEventPlacecount(event_id string) error {
	_, err := s.DB.Exec("UPDATE events SET placecount = placecount-1 WHERE id=?", event_id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetPriceIDByEventID(event_id string) (price_id string) {
	s.DB.QueryRow("SELECT priceid from events where id = ?;", event_id).Scan(&price_id)
	return
}
