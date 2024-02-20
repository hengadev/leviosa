package sqlite

import (
	"database/sql"
	"github.com/GaryHY/event-reservation-app/internal/types"
	// "github.com/google/uuid"
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
	// TODO: Add all the sql prepared statement based on some const query that you define above
}

func (s *Store) Init(queries ...string) {
	for _, query := range queries {
		_, err := s.DB.Exec(query)
		if err != nil {
			log.Fatalf("Error executing '%s' : - %s", query, err)
		}
	}
}

func (s *Store) IsAdmin(session_id string) bool {
	var role string
	query := `
        SELECT role FROM users WHERE email = 
        (SELECT email FROM sessions WHERE id = ?);
    `
	err := s.DB.QueryRow(query).Scan(&role)
	if err != nil {
		log.Fatalf("Failed to find if the user refered to the sessions id %q is an admin user", session_id)
	}
	return types.ConvertToRole(role) == types.ADMIN
}

func (s *Store) GetEventByID(id string) (event *types.Event) {
	if err := s.DB.QueryRow("SELECT * FROM events WHERE id = ?;", id).Scan(&event); err != nil {
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

func (s *Store) UpdateEvent(event *types.Event) {
	_, err := s.DB.Exec("UPDATE events set name=? where id=?;", event.Id, event.Id)
	if err != nil {
		log.Fatalf("Could not update event with id %q - %s", event.Id, err)
	}
}

func (s *Store) DeleteEvent(event *types.Event) {
	_, err := s.DB.Exec("DELETE from events where id=?;", event.Id)
	if err != nil {
		log.Fatalf("Could not delete event with id %q - %s", event.Id, err)
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

func (s *Store) GetUserId(user_email string) (id string) {
	s.DB.QueryRow("SELECT id from users where email = ?;", user_email).Scan(&id)
	return
}

// TODO: Change that function once all the field are fine !
func (s *Store) CreateUser(newUser *types.UserStored, isAdmin bool) error {
	hashpassword := hashPassword(newUser.Password)
	var role types.Role
	if isAdmin {
		role = types.ConvertToRole(newUser.Role)
	} else {
		role = types.ADMIN
	}
	_, err := s.DB.Exec("INSERT INTO users (id, email, hashpassword, role, lastname, firstname, gender, birthdate, telephone, address, city, postalcard) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);", newUser.Id, newUser.Email, hashpassword, role, newUser.LastName, newUser.FirstName, newUser.Gender, newUser.BirthDate, newUser.Telephone, newUser.Address, newUser.City, newUser.PostalCard)
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
	created_at_formatted := newSession.Created_at.Format(time.RFC822)
	_, err := s.DB.Exec("INSERT INTO sessions (id, userid, created_at) VALUES (?, ?, ?);", newSession.Id, newSession.UserId, created_at_formatted)
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
