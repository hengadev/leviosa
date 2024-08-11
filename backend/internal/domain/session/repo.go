package session

import (
	"context"
)

type Reader interface {
	GetSessionIDByUserID(ctx context.Context, userID int) (string, error)
	FindSessionByID(ctx context.Context, sessionID string) (*Session, error)
}

type Writer interface {
	CreateSession(ctx context.Context, session *Session) (string, error)
	Signout(ctx context.Context, userID int) error
}

type ReadWriter interface {
	Reader
	Writer
}
