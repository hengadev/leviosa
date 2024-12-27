package voteRepository

import (
	"context"
	"errors"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// Function to create vote for a user in a specific month and year
func (v *repository) CreateVote(ctx context.Context, userID string, days string, month, year int) error {
	query := "INSERT INTO votes (userid, days, month, year) VALUES (?, ?, ?, ?);"
	result, err := v.DB.ExecContext(ctx, query, userID, days, month, year)
	if err != nil {
		switch {

		}
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	if rowsAffected == 0 {
		return rp.NewNotCreatedErr(errors.New("no rows affected by insertion statement"), "vote")
	}
	return nil
}
