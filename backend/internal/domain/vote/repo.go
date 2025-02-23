package vote

import "context"

type Reader interface {
	FindVotesByUserID(ctx context.Context, month string, year int, userID string) (string, error)
	FindVotes(ctx context.Context, month, year int, userID string) (string, error)
	GetNextVotes(ctx context.Context, month, year int) ([]*AvailableVote, error)
	HasVote(ctx context.Context, month, year int, userID string) error
}
type Writer interface {
	CreateVote(ctx context.Context, userID string, days string, month, year int) error
	RemoveVote(ctx context.Context, userID string, month, year int) error
}

type ReadWriter interface {
	Reader
	Writer
}
