package userRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/user"
	"github.com/GaryHY/leviosa/tests/utils/factories"

	"github.com/GaryHY/test-assert"
)

func TestDeleteUser(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		userID  string
		wantErr bool
		version int64
		name    string
	}{
		{userID: factories.Johndoe.ID, wantErr: true, version: 20240811085134, name: "user not in the database"},
		{userID: "95832", wantErr: true, version: 20240811140841, name: "wrong query"}, {userID: factories.Johndoe.ID, wantErr: false, version: 20240811140841, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, userRepository.New)
			defer teardown()
			err := repo.DeleteUser(ctx, tt.userID)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
