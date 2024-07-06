package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/pkg/sqliteutil"
)

type VoteRepository struct {
	DB *sql.DB
}

func NewVoteRepository(ctx context.Context) (*VoteRepository, error) {
	connStr := os.Getenv("votedb")
	db, err := sqliteutil.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}
	// TODO: initialise the admin if the env variable is set to dev.
	// or maybe us flags for this ?
	if os.Getenv("env") == "dev" {
		ProdInit(db)
	}
	return &VoteRepository{db}, nil
}

func (v *VoteRepository) FindVotesByUserID(ctx context.Context, month, year, userID string) (string, error) {
	var votes string
	tableName := fmt.Sprintf("votes_%s_%s", month, year)
	query := fmt.Sprintf("SELECT * FROM %s WHERE userid=?;", tableName)
	if err := v.DB.QueryRowContext(ctx, query).Scan(&votes); err != nil {
		return "", rp.NewNotFoundError(err)
	}
	return votes, nil // votes the string thing
}

// TODO: I need the next votes, the past votes, the closest vote
func (v *VoteRepository) GetNextVotes(ctx context.Context) ([]*vote.Vote, error) {
	now := time.Now().UTC()
	month := int(now.Month())
	year := int(now.Year())
	var votes = make([]*vote.Vote, 8)
	condition := fmt.Sprintf("(year=? AND month>?) OR year=?", year, month+1, year+1)
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

func (v *VoteRepository) HasSession(ctx context.Context, userID, month, year int) (bool, error) {
	var res bool
	tablename := fmt.Sprintf("votes_%d_%d", month, year)
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE userid=?;", tablename)
	err := v.DB.QueryRowContext(ctx, query).Scan(&res)
	if err == sql.ErrNoRows {
		return false, rp.NewNotFoundError(err)
	}
	if err != nil {
		return false, rp.NewBadQueryErr(err)
	}
	return true, nil
}

// Function to create vote for a user in a specific month and year
func (v *VoteRepository) CreateVote(ctx context.Context, userID, days string, month, year int) error {
	tablename := fmt.Sprintf("votes_%d_%d", month, year)
	query := fmt.Sprintf("INSERT INTO %s (userid, days) VALUES (?, ?);", tablename)
	_, err := v.DB.ExecContext(ctx, query, userID, days)
	if err != nil {
		return rp.NewRessourceCreationErr(err)
	}
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
