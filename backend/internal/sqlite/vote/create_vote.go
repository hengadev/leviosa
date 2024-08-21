package voteRepository

import (
	"context"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// Function to create vote for a user in a specific month and year
func (v *VoteRepository) CreateVote(ctx context.Context, userID int, days string, month, year int) (int, error) {
	fail := func(err error) (int, error) {
		return 0, rp.NewRessourceCreationErr(err)
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
	return int(lastInsertID), nil
}
