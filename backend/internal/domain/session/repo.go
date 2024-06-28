package session

import (
	"context"
)

type Reader interface {
	GetSessionIDByUserID(ctx context.Context, userID string) (string, error)
}
type Writer interface {
	CreateSession(ctx context.Context, session *Session) (string, error)
}

type ReadWriter interface {
	Reader
	Writer
}
