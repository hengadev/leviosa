package voteRepository_test

import (
	"context"
	"testing"

	voteRepository "github.com/GaryHY/event-reservation-app/internal/sqlite/vote"
	testdb "github.com/GaryHY/event-reservation-app/pkg/sqliteutil/testdatabase"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestRemoveVote(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		userID               int
		month                int
		year                 int
		expectedRowsAffected int
		wantErr              bool
		version              int64
		name                 string
	}{
		{userID: 1, month: 4, year: 2025, expectedRowsAffected: 0, wantErr: false, version: 20240820223653, name: "no vote in database to remove"},
		{userID: 447349, month: 4, year: 2025, expectedRowsAffected: 0, wantErr: false, version: 20240820225713, name: "wrong query"},
		{userID: 1, month: 4, year: 2025, expectedRowsAffected: 1, wantErr: false, version: 20240820225713, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, err := testdb.SetupRepo(ctx, tt.version, voteRepository.New)
			defer testdb.Teardown(repo.DB)
			if err != nil {
				t.Errorf("setup repo: %s", err)
			}
			rowsAffected, err := repo.RemoveVote(ctx, tt.userID, tt.month, tt.year)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, rowsAffected, tt.expectedRowsAffected)
		})
	}
}
