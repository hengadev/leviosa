package userRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/user"
	"github.com/GaryHY/leviosa/tests/utils/factories"

	"github.com/GaryHY/test-assert"
)

func TestModifyAccount(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")

	changes := map[string]any{"FirstName": "Jane", "Gender": "F"}
	whereMap := map[string]any{"id": factories.Johndoe.ID}
	modifiedUser, err := domain.CreateWithZeroFieldModifiedObject(*factories.Johndoe, changes)
	if err != nil {
		t.Error("Failed to create object with modified field")
	}

	tests := []struct {
		userModified *models.User
		wantErr      bool
		version      int64
		name         string
	}{
		{userModified: nil, wantErr: true, version: 20240811085134, name: "nil user"},
		{userModified: factories.Johndoe, wantErr: true, version: 20240811140841, name: "user with prohibited fields for modification"},
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
