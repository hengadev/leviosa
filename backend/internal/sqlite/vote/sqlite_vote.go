package vote_repo

import (
	"database/sql"
	"log"

	"github.com/GaryHY/event-reservation-app/internal/types"
)

func (s *Store) CreateVote(newVote *types.Vote) error {
	_, err := s.DB.Exec("INSERT INTO votes (id, userid, eventid) VALUES (?, ?, ?);", newVote.Id, newVote.UserId, newVote.EventId)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) CheckVote(userId, eventId *string) bool {
	var value int
	err := s.DB.QueryRow("SELECT 1 FROM votes WHERE userid=? AND eventid=?;", userId, eventId).Scan(&value)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		log.Fatal("Cannot query if the vote already exist", err)
	}
	return true
}

func (s *Store) CheckVoteById(voteId *string) bool {
	var value int
	err := s.DB.QueryRow("SELECT 1 FROM votes WHERE id=?;", voteId).Scan(&value)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		log.Fatal("Cannot query if the vote already exist", err)
	}
	return true
}

func (s *Store) DeleteVote(voteId *string) error {
	_, err := s.DB.Exec("DELETE from votes where id=?;", voteId)
	if err != nil {
		return err
	}
	return nil
}
