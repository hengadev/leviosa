package vote_test

import (
	"context"
	"fmt"

	"github.com/GaryHY/leviosa/internal/domain/vote"
)

type MockDBKey struct {
	userID string
	month  int
	year   int
}

type MockDB map[MockDBKey]string

type StubVoteRepository struct {
	votes MockDB
}

func NewStubVoteRepository(context.Context) *StubVoteRepository {
	votes := make(map[MockDBKey]string)
	return &StubVoteRepository{votes: votes}
}

func (s *StubVoteRepository) HasVote(ctx context.Context, month, year int, userID string) error {
	for key := range s.votes {
		if key.userID == userID && key.month == month && key.year == year {
			return fmt.Errorf("vote not found")
		}
	}
	return nil
}

func (s *StubVoteRepository) CreateVote(ctx context.Context, userID string, days string, month, year int) error {
	key := MockDBKey{
		userID: userID,
		month:  month,
		year:   year,
	}
	if _, ok := s.votes[key]; ok {
		return fmt.Errorf("value in database")
	}
	s.votes[key] = days
	return nil
}

func (s *StubVoteRepository) RemoveVote(ctx context.Context, userID string, month, year int) error {
	key := MockDBKey{
		userID: userID,
		month:  month,
		year:   year,
	}
	delete(s.votes, key)
	return nil
}

func (s *StubVoteRepository) FindVotesByUserID(ctx context.Context, month string, year int, userID string) (string, error) {
	return "", nil
}

func (s *StubVoteRepository) FindVotes(ctx context.Context, month, year int, userID string) (string, error) {
	return "", nil
}

func (s *StubVoteRepository) GetNextVotes(ctx context.Context, month, year int) ([]*vote.AvailableVote, error) {
	return nil, nil
}
