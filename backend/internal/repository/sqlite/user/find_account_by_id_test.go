package userRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/user"
	"github.com/GaryHY/leviosa/tests/utils/factories"

	"github.com/GaryHY/test-assert"
)

func TestFindAccountByID(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		expectedUser *models.User
		wantErr      bool
		version      int64
		name         string
	}{
		{expectedUser: nil, wantErr: true, version: 20240811085134, name: "user not in the database"},
		{expectedUser: factories.Johndoe, wantErr: false, version: 20240811140841, name: "nominal case, user in database"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, userRepository.New)
			defer teardown()
			user, err := repo.FindAccountByID(ctx, "1")
			assert.Equal(t, err != nil, tt.wantErr)
			if tt.expectedUser != nil {
				fields := []string{"ID", "Email", "Role", "BirthDate", "LastName", "FirstName", "Gender", "Telephone", "Address", "City", "PostalCard"}
				assert.FieldsEqual(t, user, tt.expectedUser, fields)
			}
		})
	}
}
