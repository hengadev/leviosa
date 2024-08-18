package vote

import (
	"context"
	"fmt"
	"strings"

	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (s *Service) CreateVote(ctx context.Context, votes []*Vote) error {
	for _, vote := range votes {
		if pbms := vote.Valid(ctx); len(pbms) > 0 {
			return serverutil.FormatError(pbms, "vote")
		}
	}

	month := votes[0].Month
	year := votes[0].Year
	userID := votes[0].UserID

	days := formatVote(votes)
	// check if has vote
	hasVote, err := s.Repo.HasVote(ctx, month, year, userID)
	if err != nil {
		return fmt.Errorf("check if user votes %w", err)
	}
	// remove previous vote
	if hasVote {
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

// formatVote takes an array of Vote type to return formatted days respecting the order of Votes (indicating user's preference) to write in database. The separator string is a package constant.
func formatVote(votes []*Vote) string {
	var daysArr = make([]string, len(votes))
	for i, vote := range votes {
		daysArr[i] = fmt.Sprintf("%d", vote.Day)
	}
	days := strings.Join(daysArr, VoteSeparator)
	return days
}
