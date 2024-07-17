package session

import (
	"context"
)

type Reader interface {
	GetSessionIDByUserID(ctx context.Context, userID string) (string, error)
	FindSessionByID(ctx context.Context, sessionID string) (*Session, error)
}

type Writer interface {
	CreateSession(ctx context.Context, session *Session) (string, error)
	Signout(ctx context.Context, userID string)
}

type ReadWriter interface {
	Reader
	Writer
}
