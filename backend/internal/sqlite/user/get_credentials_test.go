package userRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/sqlite"
	"github.com/GaryHY/event-reservation-app/internal/sqlite/user"
	"github.com/GaryHY/event-reservation-app/pkg/testutil"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestGetCredentials(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	creds := &userService.Credentials{
		Email:    testutil.Johndoe.Email,
		Password: testutil.Johndoe.Password,
	}
	tests := []struct {
		expectedUserID   int
		expectedPassword string
		expectedRole     userService.Role
		wantErr          bool
		version          int64
		name             string
	}{
		{expectedUserID: 0, expectedPassword: "", expectedRole: userService.UNKNOWN, wantErr: true, version: 20240811085134, name: "No users in database"},
		{expectedUserID: 1, expectedPassword: creds.Password, expectedRole: userService.BASIC, wantErr: false, version: 20240819182030, name: "Multiple users in the database to retrieve"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, userRepository.New)
			defer teardown()
			userID, password, role, err := repo.GetCredentials(ctx, creds)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, userID, tt.expectedUserID)
			assert.Equal(t, role, tt.expectedRole)
			// NOTE: I can do that because I did not hashed the password in migrations (no need to test that dependency)
			assert.Equal(t, password, tt.expectedPassword)
		})
	}
}
