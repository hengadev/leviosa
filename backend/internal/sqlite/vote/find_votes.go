package voteRepository

import (
	"context"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (v *VoteRepository) FindVotes(ctx context.Context, month, year, userID int) (string, error) {
	var days string
	query := "SELECT days FROM votes WHERE userid=? and month=? and year=?;"
	if err := v.DB.QueryRowContext(ctx, query, userID, month, year).Scan(&days); err != nil {
		return "", rp.NewNotFoundError(err)
	}
	return days, nil
}
