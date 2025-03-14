package voteRepository

import (
	"context"
	"database/sql"
	"errors"

	rp "github.com/hengadev/leviosa/internal/repository"
)

func (v *repository) FindVotes(ctx context.Context, month, year int, userID string) (string, error) {
	var days string
	query := "SELECT days FROM votes WHERE user_id=? and month=? and year=?;"
	err := v.DB.QueryRowContext(ctx, query, userID, month, year).Scan(&days)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return "", rp.NewNotFoundErr(err, "vote")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return "", rp.NewContextErr(err)
		default:
			return "", rp.NewDatabaseErr(err)
		}
	}
	return days, nil
}
