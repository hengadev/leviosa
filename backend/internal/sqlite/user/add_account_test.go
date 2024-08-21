package userRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/sqlite"
	userRepository "github.com/GaryHY/event-reservation-app/internal/sqlite/user"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestAddAccount(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		usr     *user.User
		wantErr bool
		version int64
		name    string
	}{
		{usr: johndoe, wantErr: true, version: 20240811140841, name: "user already exists"},
		{usr: johndoe, wantErr: false, version: 20240811085134, name: "add the user"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, userRepository.New)
			defer teardown()
			err := repo.AddAccount(ctx, tt.usr)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
