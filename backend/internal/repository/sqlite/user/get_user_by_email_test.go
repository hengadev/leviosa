package userRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/user"
	"github.com/GaryHY/leviosa/tests/assert"
)

func TestGetUserByEmail(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/test")
	ctx := context.Background()
	tests := []struct {
		name          string
		version       int64
		emailHash     string
		expectedUser  *models.User
		expectedError error
	}{
		{
			name:          "no user in database",
			version:       20240811085134,
			emailHash:     "john.doe@example.com",
			expectedUser:  nil,
			expectedError: rp.ErrNotFound,
		},
		{
			name:          "user not in database",
			version:       20240811140841,
			emailHash:     "test@example.com",
			expectedUser:  nil,
			expectedError: rp.ErrNotFound,
		},
		{
			name:      "user in database",
			version:   20240811140841,
			emailHash: "john.doe@example.com",
			expectedUser: &models.User{
				EmailHash:           "john.doe@example.com",
				PasswordHash:        "hashedpassword",
				Picture:             "picture",
				EncryptedCreatedAt:  "2025-02-03",
				EncryptedLoggedInAt: "2025-02-03",
				Role:                "basic",
				EncryptedBirthDate:  "1998-07-12",
				LastName:            "DOE",
				FirstName:           "John",
				Gender:              "M",
				Telephone:           "0123456789",
				PostalCode:          "75000",
				City:                "Paris",
				Address1:            "01 Avenue Jean DUPONT",
				Address2:            "",
				GoogleID:            "google_id",
				AppleID:             "apple_id",
			},
			expectedError: nil,
		},
	}
	for _, tt := range tests {
		repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, userRepository.New)
		defer teardown()
		user, err := repo.GetUserByEmail(ctx, tt.emailHash)
		assert.EqualError(t, err, tt.expectedError)
		assert.ReflectEqual(t, user, tt.expectedUser)
	}
}
