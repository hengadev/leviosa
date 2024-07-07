package vote

import "context"

type Reader interface {
	FindVotesByUserID(ctx context.Context, month, year, userID string) (string, error)
	GetNextVotes(ctx context.Context, month, year int) ([]*Vote, error)
	HasVote(ctx context.Context, userID string, month, year int) (bool, error)
}
type Writer interface {
	CreateVote(ctx context.Context, userID, days string, month, year int) error
	RemoveVote(ctx context.Context, userID string, month, year int) error
}

type ReadWriter interface {
	Reader
	Writer
}
