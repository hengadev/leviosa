package voteRepository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// table available_votes

// days, month, year

func (v *VoteRepository) GetNextVotes(ctx context.Context, month, year int) ([]*vote.AvailableVote, error) {
	fail := func(err error) ([]*vote.AvailableVote, error) {
		return nil, rp.NewNotFoundError(err)
	}
	var votes []*vote.AvailableVote
	condition := fmt.Sprintf("(year=%d AND month>%d) OR year=%d", year, month, year+1)
	query := fmt.Sprintf("SELECT days, month, year from available_votes where %s LIMIT 3;", condition)
	rows, err := v.DB.QueryContext(ctx, query)
	if err != nil {
		return fail(err)
	}
	defer rows.Close()
	for rows.Next() {
		fmt.Println("in the loop brother")
		var days string
		var month_db int
		var year_db int
		err := rows.Scan(
			&days,
			&month_db,
			&year_db,
		)
		if err != nil {

			return fail(err)
		}
		availableDays, err := parseDays(days)
		if err != nil {
			return fail(err)
		}
		for _, day := range availableDays {
			votes = append(votes, &vote.AvailableVote{
				Day:   day,
				Month: month_db,
				Year:  year_db,
			})
		}
	}
	return votes, nil
}

func parseDays(days string) ([]int, error) {
	var res []int
	daysStr := strings.Split(days, vote.VoteSeparator)
	for _, dayStr := range daysStr {
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			return nil, fmt.Errorf("conversion days string to int")
		}
		res = append(res, day)
	}
	return res, nil
}
