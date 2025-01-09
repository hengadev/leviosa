package voteRepository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/GaryHY/leviosa/internal/domain/vote"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

// table available_votes

// days, month, year

func (v *repository) GetNextVotes(ctx context.Context, month, year int) ([]*vote.AvailableVote, error) {
	var votes []*vote.AvailableVote
	condition := fmt.Sprintf("(year=%d AND month>%d) OR year=%d", year, month, year+1)
	query := fmt.Sprintf("SELECT days, month, year from available_votes where %s LIMIT 3;", condition)
	rows, err := v.DB.QueryContext(ctx, query)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextErr(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	defer rows.Close()
	for rows.Next() {
		var days string
		var month_db int
		var year_db int
		err := rows.Scan(
			&days,
			&month_db,
			&year_db,
		)
		if err != nil {
			return nil, rp.NewDatabaseErr(err)
		}
		availableDays, err := parseDays(days)
		if err != nil {
			return nil, rp.NewInternalErr(err)
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
