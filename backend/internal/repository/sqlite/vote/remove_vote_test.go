package voteRepository_test

import (
	"context"
	"testing"

	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/vote"
	"github.com/GaryHY/leviosa/tests/utils"
	"github.com/GaryHY/leviosa/tests/utils/factories"

	"github.com/GaryHY/test-assert"
)

func TestRemoveVote(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/test")
	userID := factories.NewBasicUser(nil).ID
	tests := []struct {
		name        string
		version     int64
		userID      string
		month       int
		year        int
		expectedErr error
	}{
		{
			name:        "no vote in database to remove",
			version:     20240820223653,
			userID:      userID,
			month:       4,
			year:        2025,
			expectedErr: rp.ErrNotDeleted,
		},
		{
			name:        "ID not in database",
			version:     20240820225713,
			userID:      test.GenerateRandomString(16),
			month:       4,
			year:        2025,
			expectedErr: rp.ErrNotDeleted,
		},
		{
			name:        "nominal case",
			version:     20240820225713,
			userID:      userID,
			month:       4,
			year:        2025,
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, voteRepository.New)
			defer teardown()
			err := repo.RemoveVote(ctx, tt.userID, tt.month, tt.year)
			assert.EqualError(t, err, tt.expectedErr)
		})
	}
}
