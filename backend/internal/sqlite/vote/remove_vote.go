package voteRepository

import (
	"context"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (v *VoteRepository) RemoveVote(ctx context.Context, userID, month, year int) (int, error) {
	fail := func(err error) (int, error) {
		return 0, rp.NewRessourceDeleteErr(err)
	}
	query := "DELETE FROM votes WHERE userid=? AND month=? AND year=?;"
	res, err := v.DB.ExecContext(ctx, query, userID, month, year)
	if err != nil {
		fail(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		fail(err)
	}
	return int(rowsAffected), nil
}
