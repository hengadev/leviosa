package session_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
)

type KVMap map[string]*session.Values

type StubSessionRepository struct {
	sessions KVMap
}

func NewStubSessionRepository(ctx context.Context, sessions KVMap) *StubSessionRepository {
	return &StubSessionRepository{sessions: sessions}
}

func (s *StubSessionRepository) FindSessionByID(ctx context.Context, sessionID string) (*session.Session, error) {
	values, ok := s.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("no session ID in database")
	}
	return &session.Session{
		ID:         sessionID,
		UserID:     values.UserID,
		Role:       values.Role,
		LoggedInAt: values.LoggedInAt,
		CreatedAt:  values.CreatedAt,
		ExpiresAt:  values.ExpiresAt,
	}, nil
}

func (s *StubSessionRepository) CreateSession(ctx context.Context, sess *session.Session) error {
	s.sessions[sess.ID] = sess.Values()
	return nil
}

func (s *StubSessionRepository) RemoveSession(ctx context.Context, sessionID string) error {
	if _, ok := s.sessions[sessionID]; !ok {
		return errors.New("id not in database")
	}
	delete(s.sessions, sessionID)
	return nil
}
