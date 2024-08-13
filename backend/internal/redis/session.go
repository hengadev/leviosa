package redis

import (
	"context"
	"encoding/json"
	"strconv"

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
func (s *SessionRepository) GetSessionIDByUserID(ctx context.Context, userID int) (string, error) {
	userIDstr := strconv.Itoa(userID)
	value, err := s.Client.Get(ctx, userIDstr).Result()
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

func (s *SessionRepository) RemoveSession(ctx context.Context, sessionID string) error {
	if err := s.Client.Del(ctx, sessionID).Err(); err != nil {
		return rp.NewRessourceCreationErr(err)
	}
	return nil
}

// TODO:
func (s *SessionRepository) FindSessionByID(ctx context.Context, sessionID string) (*session.Session, error) {
	var res session.Session
	val, err := s.Client.Get(ctx, sessionID).Result()
	if err != nil {
		return nil, rp.NewNotFoundError(err)
	}
	json.Unmarshal([]byte(val), &res)
	return &res, nil
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
	userIDstr := strconv.Itoa(userSession.UserID)
	err = s.Client.Set(ctx, userIDstr, sessionEncoded, session.SessionDuration).Err()
	if err != nil {
		return "", rp.NewRessourceCreationErr(err)
	}
	return userSession.ID, nil
}

func (s *SessionRepository) DeleteSessionBySessionID(ctx context.Context, sessionID string) error {
	return nil
}
