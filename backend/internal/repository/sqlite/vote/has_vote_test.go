package voteRepository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/vote"

	"github.com/GaryHY/test-assert"
)

func TestHasVote(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		userID       string
		month        int
		year         int
		expectedVote bool
		wantErr      bool
		version      int64
		name         string
	}{
		{userID: "1", month: 4, year: 2025, expectedVote: false, wantErr: true, version: 20240820223653, name: "no vote in db"},
		{userID: "349324", month: 4, year: 2025, expectedVote: false, wantErr: true, version: 20240820225713, name: "wrong ID"},
		{userID: "1", month: 4, year: 2025, expectedVote: true, wantErr: false, version: 20240820225713, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, voteRepository.New)
			defer teardown()
			err := repo.HasVote(ctx, tt.month, tt.year, tt.userID)
			fmt.Println("ther error is:", err)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
