package vote

import (
	"context"
	"fmt"
	"strings"
)

func (s *Service) CreateVote(ctx context.Context, votes []*Vote) error {
	month := votes[0].Month
	year := votes[0].Year
	userID := votes[0].UserID
	days := stringifyVote(votes)
	// check if has vote
	ok, err := s.Repo.HasVote(ctx, userID, month, year)
	if err != nil {
		return fmt.Errorf("know if user has votes %w", err)
	}
	// remove previous vote
	if ok {
		if err := s.Repo.RemoveVote(ctx, userID, month, year); err != nil {
			return fmt.Errorf("remove existing vote: %w", err)
		}
	}
	// create new vote with the new information
	if err := s.Repo.CreateVote(ctx, userID, days, month, year); err != nil {
		return fmt.Errorf("add vote: %w", err)
	}
	return nil
}

func stringifyVote(votes []*Vote) string {
	var daysArr = make([]string, len(votes))
	for i, vote := range votes {
		daysArr[i] = fmt.Sprintf("%d", vote.Day)
	}
	days := strings.Join(daysArr, VoteSeparator)
	return days
}
