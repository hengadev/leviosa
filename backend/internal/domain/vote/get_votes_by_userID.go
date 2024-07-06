package vote

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	// app "github.com/GaryHY/event-reservation-app/internal/domain"
)

// Function that returns the votes (order is important) for a specific user
func (s *Service) GetVotesByUserID(ctx context.Context, monthStr, yearStr, userID string) ([]*Vote, error) {
	monthInt, err := strconv.Atoi(monthStr)
	if err != nil {
		return nil, fmt.Errorf("fail to convert string month to int")
	}
	yearInt, err := strconv.Atoi(yearStr)
	if err != nil {
		return nil, fmt.Errorf("fail to convert string year to int")
	}
	month := strings.ToLower(time.Month(monthInt).String())
	votesStr, err := s.Repo.FindVotesByUserID(ctx, month, yearStr, userID)
	if err != nil {
		return nil, fmt.Errorf("get votes by userID: %w", err)
	}
	votes, err := parseVotes(votesStr, monthInt, yearInt)
	if err != nil {
		return nil, fmt.Errorf("parse votes by userID: %w", err)
	}
	return votes, nil
}

// vote du mois, the table is going to be vote_january_2024
// userID - someformatted vote thing

// two tables are made
// votes [month-year-availabledates]
// votes_april_2024

// Function that parse string stored in repository into votes.
func parseVotes(daysStr string, month, year int) ([]*Vote, error) {
	days := strings.Split(daysStr, VoteSeparator)
	var votes = make([]*Vote, len(days))
	for i, day := range days {
		day, err := strconv.Atoi(day)
		if err != nil {
			return nil, fmt.Errorf("cannot convert string day to int")
		}
		votes[i] = &Vote{
			Day:   day,
			Month: month,
			Year:  year,
		}
	}
	return nil, nil
}
