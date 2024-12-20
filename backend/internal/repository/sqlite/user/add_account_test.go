package userRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/repository/sqlite"
	"github.com/GaryHY/event-reservation-app/internal/repository/sqlite/user"
	"github.com/GaryHY/event-reservation-app/pkg/testutil"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestAddAccount(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		usr     *userService.User
		wantErr bool
		version int64
		name    string
	}{
		{usr: testutil.Johndoe, wantErr: true, version: 20240811140841, name: "user already exists"},
		{usr: testutil.Johndoe, wantErr: false, version: 20240811085134, name: "add the user"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, userRepository.New)
			defer teardown()
			err := repo.AddAccount(ctx, tt.usr)
			assert.Equal(t, err != nil, tt.wantErr)
			// if !tt.wantErr {
			// 	assert.Equal(t, int(id), testutil.Johndoe.ID)
			// }
		})
	}
}
