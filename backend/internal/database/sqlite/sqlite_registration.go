package sqlite

import (
	"database/sql"
	"log"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/types"
)

// Function that return if there is a registration for a certain user for a certain event at a certain time.
func (s *Store) CheckRegistration(registration *types.Registration) bool {
	// TODO: check if there is the entry for begin at
	var value int
	err := s.DB.QueryRow("SELECT 1 FROM ? WHERE beginAt=?;", registration.EventId, registration.BeginAt.Format(time.RFC3339)).Scan(&value)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		log.Fatal("Cannot query due to some internal error with the database - ", err)
		return false
	}
	return true
}

func (s *Store) CreateRegistration(registration *types.Registration) error {
	_, err := s.DB.Exec("INSERT INTO ? (userid, beginAt) VALUES (?, ?);", registration.EventId, registration.UserId, registration.BeginAt.Format(time.RFC3339))
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetBeginAtByEventId(eventId string) time.Time {
	var dateTemp string
	err := s.DB.QueryRow("SELECT beginAt FROM events WHERE eventid=?;", eventId).Scan(&dateTemp)
	if err != nil {
		log.Fatalf("Failed to get the beginAt field from the row corresponding to the event %s - %s", eventId, err)
	}
	res, err := time.Parse(time.RFC3339, dateTemp)
	if err != nil {
		log.Fatal("Failed to parse the time from the database - ", err)
	}
	return res
}
