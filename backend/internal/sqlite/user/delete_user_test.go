package userRepository_test

import (
	"context"
	"testing"

	userRepository "github.com/GaryHY/event-reservation-app/internal/sqlite/user"
	testdb "github.com/GaryHY/event-reservation-app/pkg/sqliteutil/testdatabase"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestDeleteUser(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		expectedRowsAffected int
		wantErr              bool
		version              int64
		name                 string
	}{
		{expectedRowsAffected: 0, wantErr: false, version: 20240811085134, name: "user not in the database"},
		{expectedRowsAffected: 1, wantErr: false, version: 20240811140841, name: "nominal case, user in database"},
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
			rowsAffected, err := repo.DeleteUser(ctx, 1)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, rowsAffected, tt.expectedRowsAffected)
		})
	}
}
