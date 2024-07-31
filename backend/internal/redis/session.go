package redis

import (
	"context"
	"encoding/json"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

// general
type SessionRepository struct {
	Client *redis.Client
}

func NewSessionRepository(ctx context.Context, client *redis.Client) (*SessionRepository, error) {
	return &SessionRepository{client}, nil
}

// reader
// A function used for authentication and authorization.
func (s *SessionRepository) GetSessionIDByUserID(ctx context.Context, userID string) (string, error) {
	value, err := s.Client.Get(ctx, userID).Result()
	if err != nil {
		return "", rp.NewNotFoundError(err)
	}
	var sessionDecoded session.Session
	err = json.Unmarshal([]byte(value), &sessionDecoded)
	if err != nil {
		return "", err
	}
	return sessionDecoded.ID, nil
}

func (s *SessionRepository) Signout(ctx context.Context, userID string) error {
	sessionID, err := s.GetSessionIDByUserID(ctx, userID)
	if err != nil {
		return rp.NewNotFoundError(err)
	}
	err = s.Client.Del(ctx, sessionID).Err()
	if err != nil {
		return rp.NewRessourceCreationErr(err)
	}
	err = s.Client.Del(ctx, userID).Err()
	if err != nil {
		return rp.NewRessourceCreationErr(err)
	}
	return nil
}

// TODO:
func (s *SessionRepository) FindSessionByID(ctx context.Context, sessionID string) (*session.Session, error) {
	return &session.Session{}, nil
}

// writer
func (s *SessionRepository) CreateSession(ctx context.Context, userSession *session.Session) (string, error) {
	sessionEncoded, err := json.Marshal(userSession)
	if err != nil {
		return "", err
	}
	err = s.Client.Set(ctx, userSession.ID, sessionEncoded, session.SessionDuration).Err()
	if err != nil {
		return "", rp.NewRessourceCreationErr(err)
	}
	err = s.Client.Set(ctx, userSession.UserID, sessionEncoded, session.SessionDuration).Err()
	if err != nil {
		return "", rp.NewRessourceCreationErr(err)
	}
	return userSession.ID, nil
}

func (s *SessionRepository) DeleteSessionBySessionID(ctx context.Context, sessionID string) error {
	return nil
}
