package voteRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/vote"
	"github.com/GaryHY/leviosa/tests/assert"
)

func TestFindVotes(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		userID       string
		month        int
		year         int
		expectedDays string
		wantErr      bool
		version      int64
		name         string
	}{
		{userID: "1", month: 4, year: 2025, expectedDays: "", wantErr: true, version: 20240820223653, name: "no vote in db"},
		{userID: "349324", month: 4, year: 2025, expectedDays: "", wantErr: true, version: 20240820225713, name: "wrong ID"},
		{userID: "1", month: 4, year: 2025, expectedDays: "23|12|6", wantErr: false, version: 20240820225713, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, voteRepository.New)
			defer teardown()
			days, err := repo.FindVotes(ctx, tt.month, tt.year, tt.userID)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, days, tt.expectedDays)
		})
	}
}
