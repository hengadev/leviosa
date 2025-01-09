package userRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/user"
	"github.com/GaryHY/leviosa/pkg/testutil"
	"github.com/GaryHY/leviosa/tests/assert"
)

func TestAddAccount(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		usr     *models.User
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
			err := repo.AddUser(ctx, tt.usr, models.Mail)
			assert.Equal(t, err != nil, tt.wantErr)
			// if !tt.wantErr {
			// 	assert.Equal(t, int(id), testutil.Johndoe.ID)
			// }
		})
	}
}
