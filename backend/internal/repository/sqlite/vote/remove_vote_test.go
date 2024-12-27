package voteRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/repository/sqlite"
	"github.com/GaryHY/event-reservation-app/internal/repository/sqlite/vote"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestRemoveVote(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		userID  string
		month   int
		year    int
		wantErr bool
		version int64
		name    string
	}{
		{userID: "1", month: 4, year: 2025, wantErr: true, version: 20240820223653, name: "no vote in database to remove"},
		{userID: "447349", month: 4, year: 2025, wantErr: true, version: 20240820225713, name: "wrong query"},
		{userID: "1", month: 4, year: 2025, wantErr: false, version: 20240820225713, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, voteRepository.New)
			defer teardown()
			err := repo.RemoveVote(ctx, tt.userID, tt.month, tt.year)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
