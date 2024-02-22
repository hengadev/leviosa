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
		if err := rows.Scan(&event.Id, &event.Location, &event.PlaceCount, &event.Date); err != nil {
			log.Fatal("Cannot scan the event - ", err)
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

	for rows.Next() {
		event := &types.Event{}
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

func (s *Store) UpdateEvent(event *types.Event) error {
	_, err := s.DB.Exec("UPDATE events SET location=?, placecount=?, date=? WHERE id=?;", event.Location, event.PlaceCount, event.Date, event.Id)
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

// Function that returns true if user in database already
func (s *Store) CheckUser(email string) bool {
	var count int
	s.DB.QueryRow("SELECT 1 FROM users WHERE email=? ;", email).Scan(&count)
	return count == 1
}

func (s *Store) CheckUserById(user_id string) bool {
	var count int
	s.DB.QueryRow("SELECT 1 FROM users WHERE id=? ;", user_id).Scan(&count)
	return count == 1
}

func (s *Store) DeleteUser(user_id string) error {
	_, err := s.DB.Exec("DELETE FROM users WHERE id = ?", user_id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) UpdateUser(user *types.UserStored) error {
	query := `
        UPDATE users SET 
		email=?,
		password=?,
		role=?,
		lastName=?,
		firstName=?,
		gender=?,
		birthDate=?,
		telephone=?,
		address=?,
		city=?,
		postalCard=?,
        WHERE id=?;
    `
	_, err := s.DB.Exec(
		query,
		user.Email,
		user.Password,
		user.Role,
		user.LastName,
		user.FirstName,
		user.Gender,
		user.BirthDate,
		user.Telephone,
		user.Address,
		user.City,
		user.PostalCard,
	)
	if err != nil {
		return err
	}
	return nil
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
func (s *Store) CreateUser(newUser *types.UserStored) error {
	hashpassword := hashPassword(newUser.Password)
	_, err := s.DB.Exec("INSERT INTO users (id, email, hashpassword, role, lastname, firstname, gender, birthdate, telephone, address, city, postalcard) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);", newUser.Id, newUser.Email, hashpassword, newUser.Role, newUser.LastName, newUser.FirstName, newUser.Gender, newUser.BirthDate, newUser.Telephone, newUser.Address, newUser.City, newUser.PostalCard)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetUserIdBySessionId(session_id string) (id string) {
	s.DB.QueryRow("SELECT userid from sessions where id = ?;", session_id).Scan(&id)
	return
}

func (s *Store) GetAllUsers() []*types.UserStored {
	users := make([]*types.UserStored, 0)
	rows, err := s.DB.Query("SELECT * FROM users;")
	if err != nil {
		log.Fatal("Cannot get events rows - ", err)
	}
	defer rows.Close()

	for rows.Next() {
		user := &types.UserStored{}
		err := rows.Scan(
			&user.Id,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.LastName,
			&user.FirstName,
			&user.Gender,
			&user.BirthDate,
			&user.Telephone,
			&user.Address,
			&user.City,
			&user.PostalCard,
		)
		if err != nil {
			log.Fatal("Cannot scan the event - ", err)
		}
		users = append(users, user)
	}
	return users
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

func (s *Store) CreateVote(newVote *types.Vote) error {
	_, err := s.DB.Exec("INSERT INTO votes (id, userid, eventid) VALUES (?, ?, ?);", newVote.Id, newVote.UserId, newVote.EventId)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DecreaseEventPlacecount(event_id string) error {
	_, err := s.DB.Exec("UPDATE events SET placecount = placecount-1 WHERE id = ?", event_id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) CheckVote(userId, eventId *string) bool {
	var value int
	err := s.DB.QueryRow("SELECT 1 FROM votes WHERE userid=? AND eventid=?;", userId, eventId).Scan(&value)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		log.Fatal("Cannot query if the vote already exist", err)
	}
	return true
}

func (s *Store) CheckVoteById(voteId *string) bool {
	var value int
	err := s.DB.QueryRow("SELECT 1 FROM votes WHERE id=?;", voteId).Scan(&value)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		log.Fatal("Cannot query if the vote already exist", err)
	}
	return true
}

func (s *Store) DeleteVote(voteId *string) error {
	_, err := s.DB.Exec("DELETE from votes where id=?;", voteId)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) Authorize(session_id string, roleMin types.Role) bool {
	var rolestr string
	query := `
        SELECT role FROM users WHERE id = 
        (SELECT userid FROM sessions WHERE id = ?);
    `
	err := s.DB.QueryRow(query, session_id).Scan(&rolestr)
	if err != nil {
		log.Fatalf("Failed to find the role of the user refered to the sessions id %q - %s)", session_id, err)
	}
	userRole := types.ConvertToRole(rolestr)

	return userRole.IsSuperior(roleMin)
}
