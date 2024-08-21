package voteRepository

import (
	"context"
	"fmt"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (v *VoteRepository) FindVotesByUserID(ctx context.Context, month string, year, userID int) (string, error) {
	var votes string
	tableName := fmt.Sprintf("votes_%s_%d", month, year)
	query := fmt.Sprintf("SELECT * FROM %s WHERE userid=?;", tableName)
	if err := v.DB.QueryRowContext(ctx, query).Scan(&votes); err != nil {
		return "", rp.NewNotFoundError(err)
	}
	return votes, nil
}
