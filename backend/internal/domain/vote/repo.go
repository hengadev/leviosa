package vote

import "context"

type Reader interface {
	FindVotesByUserID(ctx context.Context, month string, year, userID int) (string, error)
	FindVotes(ctx context.Context, month, year, userID int) (string, error)
	GetNextVotes(ctx context.Context, month, year int) ([]*AvailableVote, error)
	HasVote(ctx context.Context, month, year, userID int) error
}
type Writer interface {
	CreateVote(ctx context.Context, userID int, days string, month, year int) error
	RemoveVote(ctx context.Context, userID int, month, year int) error
}

type ReadWriter interface {
	Reader
	Writer
}
