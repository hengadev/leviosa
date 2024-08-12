package sqlite_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	testdb "github.com/GaryHY/event-reservation-app/pkg/sqliteutil/testdatabase"
	"github.com/GaryHY/event-reservation-app/tests"
)

func TestAddAccount(t *testing.T) {
	// TODO: other cases ?
	// - email unique
	// - nom prenom unique
	// - telephone unique
	t.Setenv("TEST_MIGRATION_PATH", "./migrations/tests")
	tests := []struct {
		usr            *user.User
		wantErr        bool
		expectedUserID int
		version        int64
		name           string
	}{
		{usr: johndoe, expectedUserID: 1, wantErr: false, version: 20240811085134, name: "add the user"},
		{usr: johndoe, expectedUserID: 0, wantErr: true, version: 20240811140841, name: "user already exists"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, err := setupUserRepo(ctx, tt.version)
			defer testdb.Teardown(repo.DB)
			if err != nil {
				t.Errorf("setup repo: %s", err)
			}
			userID, err := repo.AddAccount(ctx, tt.usr)
			test.Assert(t, userID, tt.expectedUserID)
			test.Assert(t, err != nil, tt.wantErr)
			if tt.wantErr == false {
				userFromDB, err := getOnlyUser(ctx, repo.DB)
				if err != nil {
					t.Errorf("user not found after addition to database: %s", err)
				}
				// TODO: see if the users are the same ?
				_ = userFromDB
			}
		})
	}
}
