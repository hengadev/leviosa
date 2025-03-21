package throttlerService

import (
	"context"
	"time"
)

type Reader interface {
	IsLocked(ctx context.Context, key string) ([]byte, error)
}

type Writer interface {
	MakeAttempt(ctx context.Context, email string, now time.Time) error
	Reset(ctx context.Context, email string) error
}

type ReadWriter interface {
	Reader
	Writer
}
