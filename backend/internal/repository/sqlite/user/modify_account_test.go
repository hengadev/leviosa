package userRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/repository/sqlite"
	"github.com/GaryHY/event-reservation-app/internal/repository/sqlite/user"
	"github.com/GaryHY/event-reservation-app/pkg/testutil"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestModifyAccount(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")

	changes := map[string]any{"FirstName": "Jane", "Gender": "F"}
	whereMap := map[string]any{"id": testutil.Johndoe.ID}
	modifiedUser, err := domain.CreateWithZeroFieldModifiedObject(*testutil.Johndoe, changes)
	if err != nil {
		t.Error("Failed to create object with modified field")
	}

	tests := []struct {
		userModified *userService.User
		wantErr      bool
		version      int64
		name         string
	}{
		{userModified: nil, wantErr: true, version: 20240811085134, name: "nil user"},
		{userModified: testutil.Johndoe, wantErr: true, version: 20240811140841, name: "user with prohibited fields for modification"},
		{userModified: modifiedUser, wantErr: false, version: 20240811140841, name: "nominal case with valid updatable user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, userRepository.New)
			defer teardown()
			err := repo.ModifyAccount(
				ctx, tt.userModified,
				whereMap,
				"Email",
				"Password",
			)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
