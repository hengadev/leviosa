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

func TestGetCredentials(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	creds := &models.UserSignIn{
		Email:    testutil.Johndoe.Email,
		Password: testutil.Johndoe.Password,
	}
	tests := []struct {
		expectedUserID   string
		expectedPassword string
		expectedRole     models.Role
		wantErr          bool
		version          int64
		name             string
	}{
		{expectedUserID: "0", expectedPassword: "", expectedRole: models.UNKNOWN, wantErr: true, version: 20240811085134, name: "No users in database"},
		{expectedUserID: "1", expectedPassword: creds.Password, expectedRole: models.BASIC, wantErr: false, version: 20240819182030, name: "Multiple users in the database to retrieve"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, userRepository.New)
			defer teardown()
			userID, role, err := repo.GetUserSessionData(ctx, creds.Email)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, userID, tt.expectedUserID)
			assert.Equal(t, role, tt.expectedRole)
		})
	}
}
