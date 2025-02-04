package userRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/user"
	"github.com/GaryHY/leviosa/pkg/testutil"

	"github.com/GaryHY/test-assert"
)

func TestAddAccount(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/test")
	tests := []struct {
		name        string
		usr         *models.User
		version     int64
		expectedErr error
	}{
		{
			name:        "user already exists",
			usr:         testutil.Johndoe,
			version:     20240811140841,
			expectedErr: nil,
		},
		// {usr: testutil.Johndoe, expectedErr: nil, version: 20240811085134, name: "add the user"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, userRepository.New)
			defer teardown()
			err := repo.AddUser(ctx, tt.usr, models.Mail)
			assert.EqualError(t, err, tt.expectedErr)
		})
	}
}
