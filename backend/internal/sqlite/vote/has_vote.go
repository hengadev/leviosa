package voteRepository

import (
	"context"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (v *repository) HasVote(ctx context.Context, month, year, userID int) error {
	var res bool
	query := "SELECT 1 FROM votes WHERE userid=? AND month=? AND year=?;"
	err := v.DB.QueryRowContext(ctx, query, userID, month, year).Scan(&res)
	if err != nil {
		return rp.NewBadQueryErr(err)
	}
	return nil
}
