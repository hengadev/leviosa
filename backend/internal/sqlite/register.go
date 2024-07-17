package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// "github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/domain/register"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

type RegisterRepository struct {
	DB *sql.DB
}

func NewRegisterRepository(ctx context.Context, db *sql.DB) *RegisterRepository {
	return &RegisterRepository{db}
}

func (v *RegisterRepository) AddRegistration(ctx context.Context, r *register.Registration, day, year int, month string) error {
	tablename := getTablename(day, year, month)
	query := fmt.Sprintf("INSERT INTO %s (userid, eventid, beginat) values (?,?,?);", tablename)
	_, err := v.DB.ExecContext(
		ctx,
		query,
		r.UserID,
		r.EventID,
		r.BeginAt,
	)
	if err != nil {
		return rp.NewRessourceCreationErr(err)
	}
	return nil
}

func (v *RegisterRepository) HasRegistration(ctx context.Context, day, year int, month, userID string) (bool, error) {
	var hasSession int
	tablename := getTablename(day, year, month)
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE userid=?", tablename)
	if err := v.DB.QueryRowContext(ctx, query, userID).Scan(&hasSession); err != nil {
		return false, rp.NewNotFoundError(err)
	}
	return hasSession > 0, nil
}

func (v *RegisterRepository) RemoveRegistration(ctx context.Context, day, year int, month string) error {
	tablename := getTablename(day, year, month)
	query := fmt.Sprintf("DELETE FROM %s WHERE userid=?", tablename)
	if _, err := v.DB.ExecContext(ctx, query); err != nil {
		return rp.NewRessourceDeleteErr(err)
	}
	return nil
}

// old api

// Function that return if there is a registration for a certain user for a certain event at a certain time.
func (v *VoteRepository) CheckRegistration(registration *register.Registration) (bool, error) {
	var value int
	query := "SELECT 1 FROM ? WHERE beginAt=?;"
	err := v.DB.QueryRow(query, registration.EventID, registration.BeginAt.Format(time.RFC3339)).Scan(&value)
	if err == sql.ErrNoRows {
		return false, rp.NewNotFoundError(err)
	}
	if err != nil {
		return false, rp.NewBadQueryErr(err)
	}
	return true, nil
}

// utils
func getTablename(day, year int, month string) string {
	return fmt.Sprintf("registration_%d_%s_%d", day, month, year)
}

//	func (s *Store) CreateRegistration(registration *types.Registration) error {
//		_, err := s.DB.Exec("INSERT INTO ? (userid, beginAt) VALUES (?, ?);", registration.EventId, registration.UserId, registration.BeginAt.Format(time.RFC3339))
//		if err != nil {
//			return err
//		}
//		return nil
//	}
//
//	func (s *Store) GetBeginAtByEventId(eventId string) time.Time {
//		var dateTemp string
//		err := s.DB.QueryRow("SELECT beginAt FROM events WHERE eventid=?;", eventId).Scan(&dateTemp)
//		if err != nil {
//			log.Fatalf("Failed to get the beginAt field from the row corresponding to the event %s - %s", eventId, err)
//		}
//		res, err := time.Parse(time.RFC3339, dateTemp)
//		if err != nil {
//			log.Fatal("Failed to parse the time from the database - ", err)
//		}
//		return res
//	}
