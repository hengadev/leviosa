package voteRepository_test

import (
	"context"
	"testing"

	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/vote"
	test "github.com/GaryHY/leviosa/tests/utils"
	"github.com/GaryHY/leviosa/tests/utils/factories"

	"github.com/GaryHY/test-assert"
)

func TestHasVote(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", test.GetSQLiteMigrationPath())
	userID := factories.NewBasicUser(nil).ID
	tests := []struct {
		name         string
		version      int64
		userID       string
		month        int
		year         int
		expectedVote bool
		expectedErr  error
	}{
		{
			name:         "no votes in database",
			version:      20240820223653,
			userID:       userID,
			month:        4,
			year:         2025,
			expectedVote: false,
			expectedErr:  rp.ErrNotFound,
		},
		{
			name:         "wrong ID",
			version:      20240820225713,
			userID:       test.GenerateRandomString(16),
			month:        4,
			year:         2025,
			expectedVote: false,
			expectedErr:  rp.ErrNotFound,
		},
		{
			name:         "nominal case",
			version:      20240820225713,
			userID:       userID,
			month:        4,
			year:         2025,
			expectedVote: true,
			expectedErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, voteRepository.New)
			defer teardown()
			err := repo.HasVote(ctx, tt.month, tt.year, tt.userID)
			assert.EqualError(t, err, tt.expectedErr)
		})
	}
}
