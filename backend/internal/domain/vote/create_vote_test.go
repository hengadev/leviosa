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
	days := []int{1, 8, 22}
	day := now.Day()
	month, year := setup(now)
	userID, normalVote := generateValidVote(now, month, year)
	key := MockDBKey{userID: userID, month: month, year: year}
	votes := getVotesFromIntDaysArr(userID, days, month, year)
	tests := []struct {
		votes   []*vote.Vote
		wantErr bool
		name    string
	}{
		{votes: []*vote.Vote{{UserID: userID, Day: 54, Month: month, Year: year}}, wantErr: true, name: "Invalid day, too small"},
		{votes: []*vote.Vote{{UserID: userID, Day: -4, Month: month, Year: year}}, wantErr: true, name: "Invalid day, too large"},
		{votes: []*vote.Vote{{UserID: userID, Day: day, Month: -3, Year: year}}, wantErr: true, name: "Invalid month, too small"},
		{votes: []*vote.Vote{{UserID: userID, Day: day, Month: 15, Year: year}}, wantErr: true, name: "Invalid month, too large"},
		{votes: []*vote.Vote{{UserID: userID, Day: day, Month: month, Year: year - 5}}, wantErr: true, name: "Invalid year"},
		{votes: votes, wantErr: false, name: "Nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo := NewStubVoteRepository(ctx)
			service := vote.NewService(repo)
			err := service.CreateVote(ctx, tt.votes)
			_, ok := repo.votes[key]
			assert.Equal(t, err != nil, tt.wantErr)
			assert.NotEqual(t, ok, tt.wantErr)
		})
	}
	t.Run("Have vote in database", func(t *testing.T) {
		ctx := context.Background()
		repo := NewStubVoteRepository(ctx)
		votes := []*vote.Vote{normalVote}
		initialValue := "12|6|37"
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
		UserID: "23323543",
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
			t.Parallel()
			got := vote.ExportedFormatVote(tt.votes)
			assert.Equal(t, got, tt.expect)
		})
	}
}
