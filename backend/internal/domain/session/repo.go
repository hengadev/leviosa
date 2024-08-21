package session

import (
	"context"
)

type Reader interface {
	FindSessionByID(ctx context.Context, sessionID string) (*Session, error)
}

type Writer interface {
	CreateSession(ctx context.Context, session *Session) error
	RemoveSession(ctx context.Context, sessionID string) error
}

type ReadWriter interface {
	Reader
	Writer
}
