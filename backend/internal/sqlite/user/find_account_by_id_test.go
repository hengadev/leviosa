package userRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	userRepository "github.com/GaryHY/event-reservation-app/internal/sqlite/user"
	testdb "github.com/GaryHY/event-reservation-app/pkg/sqliteutil/testdatabase"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestFindAccountByID(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		expectedUser *user.User
		wantErr      bool
		version      int64
		name         string
	}{
		{expectedUser: nil, wantErr: true, version: 20240811085134, name: "user not in the database"},
		{expectedUser: johndoe, wantErr: false, version: 20240811140841, name: "nominal case, user in database"},
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
			user, err := repo.FindAccountByID(ctx, 1)
			assert.Equal(t, err != nil, tt.wantErr)
			if tt.expectedUser != nil {
				fields := []string{"ID", "Email", "Role", "BirthDate", "LastName", "FirstName", "Gender", "Telephone", "Address", "City", "PostalCard"}
				compareUser(t, fields, user, johndoe)
			}
		})
	}
}
