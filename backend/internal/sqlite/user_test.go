package sqlite_test

import (
	"context"
	"fmt"
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

func TestGetAllUsers(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "./migrations/tests")

	usersList := []*user.User{johndoe, janedoe, jeandoe}
	tests := []struct {
		expectedUsers []*user.User
		wantErr       bool
		version       int64
		name          string
	}{
		{expectedUsers: []*user.User{}, wantErr: false, version: 20240811085134, name: "No users in database"},
		{expectedUsers: usersList, wantErr: false, version: 20240819182030, name: "Multiple users in the database to retrieve"},
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
			users, err := repo.GetAllUsers(ctx)
			assert.Equal(t, err != nil, tt.wantErr)
			fields := []string{}
			for i := range len(users) {
				compareUser(t, fields, users[i], tt.expectedUsers[i])
			}
		})
	}
}

func TestGetCredentials(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "./migrations/tests")
	// TEST: cases:
	// - no user
	// - user in database

	creds := &user.Credentials{
		Email:    johndoe.Email,
		Password: johndoe.Password,
	}
	tests := []struct {
		expectedUserID   int
		expectedPassword string
		expectedRole     user.Role
		wantErr          bool
		version          int64
		name             string
	}{
		{expectedUserID: 0, expectedPassword: "", expectedRole: user.UNKNOWN, wantErr: true, version: 20240811085134, name: "No users in database"},
		{expectedUserID: 1, expectedPassword: creds.Password, expectedRole: user.BASIC, wantErr: false, version: 20240819182030, name: "Multiple users in the database to retrieve"},
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
			userID, password, role, err := repo.GetCredentials(ctx, creds)
			fmt.Println("got error:", err)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, userID, tt.expectedUserID)
			assert.Equal(t, role, tt.expectedRole)
			// NOTE: I can do that because I did not hashed the password in migration (no need to test that dependency)
			assert.Equal(t, password, tt.expectedPassword)
		})
	}
}
