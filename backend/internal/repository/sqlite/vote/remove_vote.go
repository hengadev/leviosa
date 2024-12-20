package voteRepository

import (
	"context"
	"errors"
	"fmt"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (v *repository) RemoveVote(ctx context.Context, userID, month, year int) error {
	query := "DELETE FROM votes WHERE userid=? AND month=? AND year=?;"
	res, err := v.DB.ExecContext(ctx, query, userID, month, year)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextError(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	if rowsAffected == 0 {
		return rp.NewNotDeletedErr(err, fmt.Sprintf("vote for user with ID %s", userID))
	}
	return nil
}
