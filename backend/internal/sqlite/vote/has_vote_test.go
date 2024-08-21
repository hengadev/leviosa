package voteRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/sqlite"
	"github.com/GaryHY/event-reservation-app/internal/sqlite/vote"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestHasVote(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		userID       int
		month        int
		year         int
		expectedVote bool
		wantErr      bool
		version      int64
		name         string
	}{
		{userID: 1, month: 4, year: 2025, expectedVote: false, wantErr: true, version: 20240820223653, name: "no vote in db"},
		{userID: 349324, month: 4, year: 2025, expectedVote: false, wantErr: true, version: 20240820225713, name: "wrong ID"},
		{userID: 1, month: 4, year: 2025, expectedVote: true, wantErr: false, version: 20240820225713, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, voteRepository.New)
			defer teardown()
			days, err := repo.HasVote(ctx, tt.month, tt.year, tt.userID)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, days, tt.expectedVote)
		})
	}
}
