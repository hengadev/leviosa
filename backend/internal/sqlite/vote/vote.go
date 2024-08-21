package voteRepository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

type VoteRepository struct {
	DB *sql.DB
}

func (v *VoteRepository) GetDB() *sql.DB {
	return v.DB
}

func New(ctx context.Context, db *sql.DB) *VoteRepository {
	return &VoteRepository{db}
}

// TODO: I need the next votes, the past votes, the closest vote
// get inspiration from some similar function in sqlite/event.go
func (v *VoteRepository) GetNextVotes(ctx context.Context, month, year int) ([]*vote.Vote, error) {
	var votes = make([]*vote.Vote, 8)
	condition := fmt.Sprintf("(year=%d AND month>%d) OR year=%d", year, month+1, year+1)
	query := fmt.Sprintf("SELECT (month, year) from votes where %s LIMIT 8;", condition)
	rows, err := v.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, rp.NewNotFoundError(err)
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &vote.Vote{}
		err := rows.Scan(
			&tmp.Month,
			&tmp.Year,
		)
		if err != nil {
			return nil, rp.NewNotFoundError(err)
		}
		votes = append(votes, tmp)
	}
	return votes, nil
}

// TODO: implement that function
func (v *VoteRepository) RemoveVote(ctx context.Context, userID, month, year int) error {
	return nil
}

func (s *VoteRepository) CheckVote(ctx context.Context, userId, eventId *string) (bool, error) {
	var value int
	err := s.DB.QueryRowContext(ctx, "SELECT 1 FROM votes WHERE userid=? AND eventid=?;", userId, eventId).Scan(&value)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, rp.NewNotFoundError(err)
	}
	return true, nil
}

func (s *VoteRepository) CheckVoteById(ctx context.Context, voteId *string) (bool, error) {
	var value int
	err := s.DB.QueryRowContext(ctx, "SELECT 1 FROM votes WHERE id=?;", voteId).Scan(&value)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, rp.NewNotFoundError(err)
	}
	return true, nil
}

func (s *VoteRepository) DeleteVote(ctx context.Context, voteId *string) error {
	_, err := s.DB.ExecContext(ctx, "DELETE from votes where id=?;", voteId)
	if err != nil {
		return rp.NewRessourceDeleteErr(err)
	}
	return nil
}
