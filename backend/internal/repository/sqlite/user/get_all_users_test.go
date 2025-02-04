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

func TestGetAllUsers(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	usersList := []*models.User{testutil.Johndoe, testutil.Janedoe, testutil.Jeandoe}
	tests := []struct {
		expectedUsers []*models.User
		wantErr       bool
		version       int64
		name          string
	}{
		{expectedUsers: []*models.User{}, wantErr: false, version: 20240811085134, name: "No users in database"},
		{expectedUsers: usersList, wantErr: false, version: 20240819182030, name: "Multiple users in the database to retrieve"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, userRepository.New)
			defer teardown()
			users, err := repo.GetAllUsers(ctx)
			assert.Equal(t, err != nil, tt.wantErr)
			fields := []string{}
			for i := range len(users) {
				defer testutil.RecoverCompareUser()
				testutil.CompareUser(t, fields, users[i], tt.expectedUsers[i])
			}
		})
	}
}
