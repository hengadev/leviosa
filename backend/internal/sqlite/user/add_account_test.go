package userRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	userRepository "github.com/GaryHY/event-reservation-app/internal/sqlite/user"
	testdb "github.com/GaryHY/event-reservation-app/pkg/sqliteutil/testdatabase"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestAddAccount(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		usr            *user.User
		wantErr        bool
		expectedUserID int
		version        int64
		name           string
	}{
		{usr: johndoe, expectedUserID: 0, wantErr: true, version: 20240811140841, name: "user already exists"},
		{usr: johndoe, expectedUserID: 1, wantErr: false, version: 20240811085134, name: "add the user"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, err := testdb.SetupRepo(ctx, tt.version, userRepository.New)
			defer testdb.Teardown(repo.DB)
			if err != nil {
				t.Errorf("setup repo: %s", err)
			}
			userID, err := repo.AddAccount(ctx, tt.usr)
			assert.Equal(t, userID, tt.expectedUserID)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
