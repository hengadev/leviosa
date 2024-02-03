package sqlite

import (
	"database/sql"
	// "time"
	// "fmt"
	"log"

	"github.com/GaryHY/event-reservation-app/internal/types"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
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
	s.DB.QueryRow("SELECT * FROM events WHERE id = ?;", id).Scan(&event.Id, &event.Name)
	return
}

func (s *Store) PostEvent(name string) {
	// query := fmt.Sprintf("INSERT INTO events (name) VALUES ('%s');", name)
	// _, err := s.DB.Exec(query)
	_, err := s.DB.Exec("INSERT INTO events (name) VALUES (?);", name)
	if err != nil {
		log.Fatal("Could not insert new event into the database - ", err)
	}
}

// Function that returns true if user in database already
func (s *Store) CheckUser(user types.User) bool {
	var count int
	s.DB.QueryRow("SELECT COUNT(email) from users where email=? ;", user.Email).Scan(&count)
	return count == 1
}

func (s *Store) GetHashPassword(user types.User) (hashpassword string) {
	s.DB.QueryRow("SELECT hashpassword from users where email = ?;", user.Email).Scan(&hashpassword)
	return
}

func (s *Store) CreateUser(newUser types.User) error {
	hashpassword := hashPassword(newUser.Password)
	_, err := s.DB.Exec("INSERT INTO users (email, hashpassword) VALUES (?, ?);", newUser.Email, hashpassword)
	if err != nil {
		return err
	}
	return nil
}

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Fatal("Failed hashing the password - ", err)
	}
	return string(bytes)
}

func (s *Store) CreateSession(id string, newSession *types.Session) error {
	_, err := s.DB.Exec("INSERT INTO sessions (id, email, created_at, expired_at) VALUES (?, ?, ?, ?);", id, newSession.Email, newSession.Created_at, newSession.Expiry)
	if err != nil {
		return err
	}

	// TODO: Use the time.After function for auto deletion of the session
	// time.After(newSession.Expiry)

	return nil
}

// TODO: Add  the level of auth I want to verify
func (s *Store) AuthorizeUser(user types.User) bool {
	// TODO: Something that has to do with verifying if user if authorized to do some actions
	return true
}
