package redis

import (
	"context"
	"encoding/json"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	rdu "github.com/GaryHY/event-reservation-app/pkg/redisutil"

	"github.com/redis/go-redis/v9"
)

// general
type SessionRepository struct {
	Client *redis.Client
}

func NewSessionRepository(ctx context.Context, opts ...rdu.RedisOption) (*SessionRepository, error) {
	// Add the redis options with it, if there are any ?
	db, err := rdu.Connect(ctx, opts...)
	if err != nil {
		return nil, err
	}
	store := &SessionRepository{db}
	// The admin user for testing purposes.
	// the admin has the id 3439...
	// queries := make(map[string]interface{})
	// queries["session:3439434532245"] = struct {
	// 	ID     string `json:"id"`
	// 	UserID string `json:"userID"`
	// }{
	// 	"U2343r23490J4", // the session ID for the admin user.
	// 	"3439434532245",
	// }
	// rdu.Init(ctx, store.client, queries)
	return store, nil
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
	err = s.Client.Set(ctx, userSession.ID, sessionEncoded, session.SessionExpirationDuration).Err()
	if err != nil {
		return "", rp.NewRessourceCreationErr(err)
	}
	err = s.Client.Set(ctx, userSession.UserID, sessionEncoded, session.SessionExpirationDuration).Err()
	if err != nil {
		return "", rp.NewRessourceCreationErr(err)
	}
	return userSession.ID, nil
}

func (s *SessionRepository) DeleteSessionBySessionID(ctx context.Context, sessionID string) error {
	err := s.Client.Del(ctx, sessionID).Err()
	if err != nil {
		return rp.NewRessourceCreationErr(err)
	}
	return nil
}

func (s *SessionRepository) DeleteSessionByUserID(ctx context.Context, userID string) error {
	err := s.Client.Del(ctx, userID).Err()
	if err != nil {
		return rp.NewRessourceCreationErr(err)
	}
	return nil
}
