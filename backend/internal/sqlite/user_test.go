package sqlite_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/sqlite"
	testdb "github.com/GaryHY/event-reservation-app/pkg/sqliteutil/testdatabase"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestAddAccount(t *testing.T) {
	// TODO: other cases ?
	// - email unique
	// - nom prenom unique
	// - telephone unique
	t.Setenv("TEST_MIGRATION_PATH", "./migrations/tests")
	tests := []struct {
		usr            *user.User
		wantErr        bool
		expectedUserID int
		version        int64
		name           string
	}{
		{usr: johndoe, expectedUserID: 0, wantErr: true, version: 20240811140841, name: "user already exists"},
		{usr: johndoe, expectedUserID: 1, wantErr: false, version: 20240811085134, name: "add the user"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, err := setupRepo(ctx, tt.version, sqlite.NewUserRepository)
			defer testdb.Teardown(repo.DB)
			if err != nil {
				t.Errorf("setup repo: %s", err)
			}
			userID, err := repo.AddAccount(ctx, tt.usr)
			assert.Equal(t, userID, tt.expectedUserID)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}

func TestFindAccountByID(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "./migrations/tests")
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
			repo, err := setupRepo(ctx, tt.version, sqlite.NewUserRepository)
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

func TestDeleteUser(t *testing.T) {
	// TODO: test cases
	// - no user
	// - has user
	t.Setenv("TEST_MIGRATION_PATH", "./migrations/tests")
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
			repo, err := setupRepo(ctx, tt.version, sqlite.NewUserRepository)
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

func TestModifyAccount(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "./migrations/tests")

	changes := map[string]any{"FirstName": "Jane", "Gender": "F"}
	whereMap := map[string]any{"id": johndoe.ID}
	modifiedUser, err := createWithZeroFieldModifiedObject(*johndoe, changes)
	if err != nil {
		t.Error("Failed to create object with modified field")
	}

	tests := []struct {
		userModified         *user.User
		expectedRowsAffected int
		wantErr              bool
		version              int64
		name                 string
	}{
		{userModified: nil, expectedRowsAffected: 0, wantErr: true, version: 20240811085134, name: "nil user"},
		{userModified: johndoe, expectedRowsAffected: 0, wantErr: true, version: 20240811140841, name: "user with prohibited fields for modification"},
		{userModified: modifiedUser, expectedRowsAffected: 1, wantErr: false, version: 20240811140841, name: "nominal case with valid updatable user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, err := setupRepo(ctx, tt.version, sqlite.NewUserRepository)
			defer testdb.Teardown(repo.DB)
			if err != nil {
				t.Errorf("setup repo: %s", err)
			}
			rowsAffected, err := repo.ModifyAccount(
				ctx, tt.userModified,
				whereMap,
				"Email",
				"Password",
			)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, rowsAffected, tt.expectedRowsAffected)
		})
	}
}
