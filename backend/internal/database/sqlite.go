package sqlite

import (
	"database/sql"
	"fmt"
	"github.com/GaryHY/event-reservation-app/internal/types"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type Store struct {
	DB *sql.DB
}

func NewStore(connStr string) (*Store, error) {
	if connStr == "" {
		connStr = ":memory:"
	}
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		log.Fatal("Cannot connect to the database - ", err)
	}
	return &Store{db}, nil
}

func (s *Store) Init(queries ...string) {
	for _, query := range queries {
		_, err := s.DB.Exec(query)
		if err != nil {
			log.Fatalf("Error executing '%s' : - %s", query, err)
		}
	}
}

func (s *Store) GetEventNameByID(id string) (event types.Event) {
	query := fmt.Sprintf("SELECT * FROM events WHERE id=%s;", id)
	s.DB.QueryRow(query).Scan(&event.Id, &event.Name)
	return
}

func (s *Store) PostEvent(name string) {
	query := fmt.Sprintf("INSERT INTO events (name) VALUES ('%s');", name)
	_, err := s.DB.Exec(query)
	if err != nil {
		log.Fatal("Could not insert new event into the database - ", err)
	}
}

// Function that returns true if user in database already
func (s *Store) CheckUser(user types.User) bool {
	var count int
	query := fmt.Sprintf("SELECT COUNT(email) from users where email=%s;", user.Email)
	s.DB.QueryRow(query).Scan(&count)
	return count == 1
}

func (s *Store) CreateUser(newUser types.User) error {
	// TODO:
	query := fmt.Sprintf("INSERT INTO users (email, hashpassword) VALUES ('%s', '%s');", newUser.Email, newUser.Password)
	_, err := s.DB.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

// TODO: Add  the level of auth I want to verify
func (s *Store) VerifyUser(user types.User) bool {
	// TODO: Something that has to do with verifying if user if authorized to do some actions
	return true
}
