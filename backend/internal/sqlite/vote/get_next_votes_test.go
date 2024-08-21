package voteRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/vote"
	"github.com/GaryHY/event-reservation-app/internal/sqlite"
	"github.com/GaryHY/event-reservation-app/internal/sqlite/vote"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestGetNextVotes(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		month                  int
		year                   int
		expectedAvailableVotes []*vote.AvailableVote
		wantErr                bool
		version                int64
		name                   string
	}{
		{month: 4, year: 2025, expectedAvailableVotes: nil, wantErr: false, version: 20240821105317, name: "no available votes in db"},
		{month: 8, year: 2025, expectedAvailableVotes: nil, wantErr: false, version: 20240821110737, name: "wrong query, month does not exist"},
		{month: 17, year: 2025, expectedAvailableVotes: nil, wantErr: false, version: 20240821110737, name: "wrong query, month too large"},
		{month: 17, year: 1998, expectedAvailableVotes: nil, wantErr: false, version: 20240821110737, name: "wrong query, year too small"},
		{month: 3, year: 2025, expectedAvailableVotes: availableVotesArr, wantErr: false, version: 20240821110737, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, voteRepository.New)
			defer teardown()
			days, err := repo.GetNextVotes(ctx, tt.month, tt.year)
			assert.Equal(t, err != nil, tt.wantErr)
			for i, day := range days {
				assert.ReflectEqual(t, day, tt.expectedAvailableVotes[i])
			}
		})
	}
}

func TestParseDays(t *testing.T) {
	tests := []struct {
		days         string
		expectedDays []int
		wantErr      bool
		name         string
	}{
		{days: "", expectedDays: nil, wantErr: true, name: "no days provided"},
		{days: "12|24|28", expectedDays: []int{12, 24, 28}, wantErr: false, name: "nominal case"},
		{days: "1", expectedDays: []int{1}, wantErr: false, name: "just one day"},
		{days: "rerger", expectedDays: nil, wantErr: true, name: "wrong days format"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			days, err := voteRepository.ExportedParseDays(tt.days)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.ReflectEqual(t, days, tt.expectedDays)
		})
	}
}
