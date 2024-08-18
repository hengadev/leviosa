package vote_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestCreateVote(t *testing.T) {
	now := time.Now().UTC()
	userID := 45021934
	days, month, year := setupDateComponent(now)
	normalVote := &vote.Vote{
		UserID: userID,
		Day:    now.Day(),
		Month:  month,
		Year:   year,
	}
	key := MockDBKey{userID: userID, month: month, year: year}
	votes := getVotes(days, *normalVote)
	dayVote, monthVote, yearVote := getInvalidVotes(*normalVote)
	tests := []struct {
		votes   []*vote.Vote
		wantErr bool
		name    string
	}{
		{votes: []*vote.Vote{dayVote, normalVote}, wantErr: true, name: "Invalid days"},
		{votes: []*vote.Vote{monthVote, normalVote}, wantErr: true, name: "Invalid months"},
		{votes: []*vote.Vote{yearVote, normalVote}, wantErr: true, name: "Invalid year"},
		{votes: votes, wantErr: false, name: "Nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := NewStubVoteRepository(ctx)
			service := vote.NewService(repo)
			err := service.CreateVote(ctx, tt.votes)
			fmt.Println("the error that I get is :", err)
			_, ok := repo.votes[key]
			assert.Equal(t, err != nil, tt.wantErr)
			assert.NotEqual(t, ok, tt.wantErr)
		})
	}
	t.Run("Have vote in database", func(t *testing.T) {
		ctx := context.Background()
		repo := NewStubVoteRepository(ctx)
		votes := []*vote.Vote{normalVote}
		initialValue := "some value that should not be there"
		repo.votes[key] = initialValue
		service := vote.NewService(repo)
		err := service.CreateVote(ctx, votes)
		got, ok := repo.votes[key]
		assert.Equal(t, err != nil, false)
		assert.Equal(t, ok, true)
		assert.NotEqual(t, got, initialValue)
	})
}

func TestFormatVote(t *testing.T) {
	now := time.Now()
	days := []int{12, 4, 5, 30}
	randVote := &vote.Vote{
		UserID: 23323543,
		Day:    now.Day(),
		Month:  int(now.Month()),
		Year:   now.Year(),
	}
	var expect string
	var votes []*vote.Vote
	for _, day := range days {
		votes = append(votes, &vote.Vote{
			UserID: randVote.UserID,
			Day:    day,
			Month:  randVote.Month,
			Year:   randVote.Year,
		})
		expect += fmt.Sprintf("%d%s", day, vote.VoteSeparator)
	}
	expect = expect[:len(expect)-1]
	tests := []struct {
		votes  []*vote.Vote
		expect string
		name   string
	}{
		{votes: []*vote.Vote{}, expect: "", name: "No vote provided"},
		{votes: []*vote.Vote{randVote}, expect: fmt.Sprintf("%d", randVote.Day), name: "One vote provided"},
		{votes: votes, expect: expect, name: "Multiple votes provided"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := vote.ExportFormatVote(tt.votes)
			assert.Equal(t, got, tt.expect)
		})
	}
}

func getVotes(days []int, normalVote vote.Vote) []*vote.Vote {
	var votes []*vote.Vote
	for _, day := range days {
		cloneVote := normalVote
		cloneVote.Day = day
		votes = append(votes, &cloneVote)
	}
	return votes
}

func getInvalidVotes(normalVote vote.Vote) (*vote.Vote, *vote.Vote, *vote.Vote) {
	dayVote := normalVote
	monthVote := normalVote
	yearVote := normalVote
	dayVote.Day = 50
	monthVote.Month = 13
	yearVote.Year = normalVote.Year - 1
	return &dayVote, &monthVote, &yearVote
}

func setupDateComponent(now time.Time) ([]int, int, int) {
	days := []int{1, 8, 22}
	var voteMonth int
	if now.Month() == 12 {
		voteMonth = 1
	} else {
		voteMonth = int(now.Month()) + 1
	}
	year := now.Year()
	return days, voteMonth, year
}
