package vote_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/GaryHY/leviosa/internal/domain/vote"

	"github.com/GaryHY/test-assert"
)

func TestGetVotesByUserID(t *testing.T) {
	userID := "94435302"
	now := time.Now().UTC()
	month, year := setup(now)

	monthStr := strconv.Itoa(month)
	invalidMonthConvert := time.Month(12).String()
	invalidYearConvert := "this year"

	yearStr := strconv.Itoa(year)

	days := []int{2, 13, 17, 26}
	expectedVotes := getVotesFromIntDaysArr(userID, days, month, year)
	formattedDays := getFormattedDayFromIntArr(days)

	key := MockDBKey{userID: userID, month: month, year: year}
	tests := []struct {
		month         string
		year          string
		expectedVotes []*vote.Vote
		wantErr       bool
		name          string
	}{
		{month: monthStr, year: yearStr, expectedVotes: nil, wantErr: false, name: "No votes in database for user"},
		{month: invalidMonthConvert, year: yearStr, expectedVotes: nil, wantErr: true, name: "Invalid month, string to int conversion error"},
		{month: invalidYearConvert, year: yearStr, expectedVotes: nil, wantErr: true, name: "Invalid year, string to int conversion error"},
		{month: monthStr, year: yearStr, expectedVotes: expectedVotes, wantErr: false, name: "Nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo := NewStubVoteRepository(ctx)
			if tt.expectedVotes != nil {
				repo.votes[key] = formattedDays
			}
			service := vote.NewService(repo)
			votes, err := service.GetVotesByUserID(ctx, tt.month, tt.year, userID)
			assert.Equal(t, err != nil, tt.wantErr)
			for i := range len(votes) {
				assert.ReflectEqual(t, votes[i], tt.expectedVotes[i])
			}
		})
	}
}

// func TestParseVotes(t *testing.T) {
// 	now := time.Now().UTC()
// 	month, year := setup(now)
// 	days := "3|12|25"
// 	tests := []struct {
// 		name          string
// 		days          string
// 		month         int
// 		year          int
// 		expectedVotes []*vote.Vote
// 		expectErr     bool
// 	}{
// 		{
// 			name:          "No days provided",
// 			days:          "",
// 			month:         month,
// 			year:          year,
// 			expectedVotes: nil,
// 			expectErr:     false,
// 		},
// 		{
// 			name:          "Invalid days contain string non convertible to int",
// 			days:          "a|12|23",
// 			month:         month,
// 			year:          year,
// 			expectedVotes: nil,
// 			expectErr:     true,
// 		},
// 		{
// 			name:          "Invalid day",
// 			days:          "3|56|25",
// 			month:         month,
// 			year:          year,
// 			expectedVotes: nil,
// 			expectErr:     true,
// 		},
// 		{
// 			name:          "Invalid month",
// 			days:          days,
// 			month:         13,
// 			year:          year,
// 			expectedVotes: nil,
// 			expectErr:     true,
// 		},
// 		{
// 			name:          "Invalid year",
// 			days:          days,
// 			month:         month,
// 			year:          year - 5,
// 			expectedVotes: nil,
// 			expectErr:     true,
// 		},
// 		{
// 			days: days,
// 			expectedVotes: []*vote.Vote{
// 				{Day: 3, Month: month, Year: year},
// 				{Day: 12, Month: month, Year: year},
// 				{Day: 25, Month: month, Year: year},
// 			},
// 			expectErr: false, month: month, year: year, name: "Nominal case"},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()
// 			ctx := context.Background()
// 			votes, err := vote.ExportedParseVotes(ctx, tt.days, tt.month, tt.year)
// 			assert.Equal(t, err != nil, tt.expectErr)
// 			for i := range len(votes) {
// 				assert.ReflectEqual(t, votes[i], tt.expectedVotes[i])
// 			}
// 		})
// 	}
// }
