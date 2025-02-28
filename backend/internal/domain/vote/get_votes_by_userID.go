package vote

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hengadev/leviosa/pkg/errsx"
)

// Function that returns the votes (order is important) for a specific user
func (s *Service) GetVotesByUserID(ctx context.Context, monthStr, yearStr string, userID string) ([]*Vote, error) {
	monthInt, err := strconv.Atoi(monthStr)
	if err != nil {
		return nil, fmt.Errorf("fail to convert string month to int")
	}
	yearInt, err := strconv.Atoi(yearStr)
	if err != nil {
		return nil, fmt.Errorf("fail to convert string year to int")
	}
	votesStr, err := s.Repo.FindVotesByUserID(ctx, monthStr, yearInt, userID)
	if err != nil {
		return nil, fmt.Errorf("get votes by userID: %w", err)
	}
	votes, pbms := parseVotes(ctx, votesStr, monthInt, yearInt)
	if len(pbms) > 0 {
		return nil, fmt.Errorf("parse votes by userID: %w", pbms)
	}
	return votes, nil
}

// vote du mois, the table is going to be vote_january_2024
// userID - someformatted vote thing

// two tables are made
// votes [month-year-availabledates]
// votes_april_2024

// parseVotes parses string stored in database into votes.
func parseVotes(ctx context.Context, daysStr string, month, year int) ([]*Vote, errsx.Map) {
	var pbms errsx.Map
	if daysStr == "" {
		return nil, nil
	}
	days := strings.Split(daysStr, VoteSeparator)
	var votes = make([]*Vote, len(days))
	for i, day := range days {
		day, err := strconv.Atoi(day)
		if err != nil {
			pbms.Set("convert string day to int", err)
		}
		vote := &Vote{Day: day, Month: month, Year: year}
		if validPbms := vote.Valid(ctx); len(validPbms) > 0 {
			pbms.Set("vote validation", validPbms.Error())
		}
		votes[i] = vote
	}
	return votes, pbms
}
