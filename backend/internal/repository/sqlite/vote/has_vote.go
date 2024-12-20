package voteRepository

import (
	"context"
	"database/sql"
	"errors"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (v *repository) HasVote(ctx context.Context, month, year, userID int) error {
	var res bool
	query := "SELECT 1 FROM votes WHERE userid=? AND month=? AND year=?;"
	err := v.DB.QueryRowContext(ctx, query, userID, month, year).Scan(&res)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return rp.NewNotFoundError(err, "votes")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextError(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}
	return nil
}
