package userRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/sqlite"
	"github.com/GaryHY/event-reservation-app/internal/sqlite/user"
	"github.com/GaryHY/event-reservation-app/pkg/testutil"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestDeleteUser(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		userID               int
		expectedRowsAffected int
		wantErr              bool
		version              int64
		name                 string
	}{
		{userID: testutil.Johndoe.ID, expectedRowsAffected: 0, wantErr: false, version: 20240811085134, name: "user not in the database"},
		{userID: 95832, expectedRowsAffected: 0, wantErr: false, version: 20240811140841, name: "wrong query"},
		{userID: testutil.Johndoe.ID, expectedRowsAffected: 1, wantErr: false, version: 20240811140841, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, userRepository.New)
			defer teardown()
			rowsAffected, err := repo.DeleteUser(ctx, tt.userID)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, rowsAffected, tt.expectedRowsAffected)
		})
	}
}
