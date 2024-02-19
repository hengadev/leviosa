package sqlite

import (
	"database/sql"
	"github.com/GaryHY/event-reservation-app/internal/types"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
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

func (s *Store) GetEventByID(id string) (event types.Event) {
	if err := s.DB.QueryRow("SELECT * FROM events WHERE id = ?;", id).Scan(&event.Id); err != nil {
		log.Fatalf("Error getting event with id '%s' : - %s", id, err)
	}
	return
}

func (s *Store) GetAllEvents() []types.Event {
	events := []types.Event{}
	rows, err := s.DB.Query("SELECT * FROM events;")
	if err != nil {
		log.Fatal("Cannot get events rows - ", err)
	}
	defer rows.Close()

	for rows.Next() {
		var event types.Event
		if err := rows.Scan(&event.Id, &event.Location, &event.PlaceCount, &event.Date); err != nil {
			log.Fatal("Cannot scan the event - ", err)
		}
		events = append(events, event)
	}
	return events
}

func (s *Store) PostEvent(event *types.Event) {
	_, err := s.DB.Exec("INSERT INTO events (id, location, placecount, date) VALUES (?, ?, ?, ?)", event.Id, event.Location, event.PlaceCount, event.Date)
	if err != nil {
		log.Fatal("Could not insert new event into the database - ", err)
	}
}

// Function that returns true if user in database already
func (s *Store) CheckUser(email string) bool {
	var count int
	s.DB.QueryRow("SELECT COUNT(email) from users where email=? ;", email).Scan(&count)
	return count == 1
}

func (s *Store) GetHashPassword(user *types.User) (hashpassword string) {
	s.DB.QueryRow("SELECT hashpassword from users where email = ?;", user.Email).Scan(&hashpassword)
	return
}

// TODO: Change that function once all the field are fine !
func (s *Store) CreateUser(newUser *types.User) error {
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

func (s *Store) CreateSession(newSession *types.Session) error {
	// TODO: format the time before writting in the database
	created_at_formatted := newSession.Created_at.Format(time.RFC822)
	_, err := s.DB.Exec("INSERT INTO sessions (id, email, created_at) VALUES (?, ?, ?);", newSession.Id, newSession.Email, created_at_formatted)
	if err != nil {
		return err
	}
	return nil
}

// TODO: Seems like this function is obsolete ?
func (s *Store) DeleteSession(session *types.Session) error {
	_, err := s.DB.Exec("DELETE FROM sessions WHERE id=?;", session.Id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteSessionByID(id string) error {
	_, err := s.DB.Exec("DELETE FROM sessions WHERE id=?;", id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) HasSession(id string) bool {
	var got_id string
	s.DB.QueryRow("SELECT id FROM sessions WHERE id=?;", id).Scan(&got_id)
	if got_id == "" {
		return false
	} else {
		created_at_parsed := s.parseCreatedAt(id)
		if created_at_parsed.Add(types.SessionDuration).Before(time.Now()) {
			if err := s.DeleteSessionByID(id); err != nil {
				log.Fatal("Failed to deleted from session")
			}
			return false
		}
		return true
	}
}

func (s *Store) parseCreatedAt(id string) (created_at_parsed time.Time) {
	var created_at string
	s.DB.QueryRow("SELECT created_at FROM sessions WHERE id=?;", id).Scan(&created_at)
	created_at_parsed, err := time.Parse(time.RFC822, created_at)
	if err != nil {
		log.Fatal("Failed parsing date - ", err)
	}
	return
}

// TODO: Add  the level of auth I want to verify
func (s *Store) AuthorizeUser(user types.User) bool {
	// TODO: Something that has to do with verifying if user if authorized to do some actions
	return true
}
