package voteRepository

import (
	"context"
	"fmt"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// Function to create vote for a user in a specific month and year
func (v *repository) CreateVote(ctx context.Context, userID int, days string, month, year int) error {
	fail := func(err error) error {
		return rp.NewNotCreatedErr(err)
	}
	query := "INSERT INTO votes (userid, days, month, year) VALUES (?, ?, ?, ?);"
	res, err := v.DB.ExecContext(ctx, query, userID, days, month, year)
	if err != nil {
		return fail(err)
	}
	lastInsertID, err := res.LastInsertId()
	if err != nil {
		return fail(err)
	}
	if lastInsertID == 0 {
		return fmt.Errorf("vote not found")
	}
	return nil
}
