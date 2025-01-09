package voteRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/vote"
	"github.com/GaryHY/leviosa/tests/assert"
)

func TestCreateVote(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		userID               string
		days                 string
		month                int
		year                 int
		expectedLastInsertID int
		wantErr              bool
		version              int64
		name                 string
	}{
		{userID: "1", days: "5|18|26", month: 4, year: 2025, expectedLastInsertID: 0, wantErr: true, version: 20240820225713, name: "vote already exists"},
		{userID: "1", days: "", month: 5, year: 2025, expectedLastInsertID: 0, wantErr: true, version: 20240820225713, name: "check non empty days constraint for days"},
		{userID: "1", days: "23|12|6", month: 0, year: 2025, expectedLastInsertID: 0, wantErr: true, version: 20240820223653, name: "month to small"},
		{userID: "1", days: "23|12|6", month: 16, year: 2025, expectedLastInsertID: 0, wantErr: true, version: 20240820223653, name: "month too large"},
		{userID: "1", days: "23|12|6", month: 4, year: 1998, expectedLastInsertID: 0, wantErr: true, version: 20240820223653, name: "year too small"},
		{userID: "1", days: "23|12|6", month: 4, year: 2025, expectedLastInsertID: 1, wantErr: false, version: 20240820223653, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, voteRepository.New)
			defer teardown()
			err := repo.CreateVote(ctx, tt.userID, tt.days, tt.month, tt.year)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
