package sqlite

import (
	"log"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/types"
)

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
		if created_at_parsed.Add(types.CookieDuration).Before(time.Now()) {
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
