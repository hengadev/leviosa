package userRepository

import (
	"context"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
)

// AddUser adds a user to the 'users' table, considering the specified provider type.
// If accounts from other providers exist, it ensures proper account linking.
func (u *Repository) AddUser(ctx context.Context, user *models.User, provider models.ProviderType) error {
	if err := u.addGenericUser(ctx, user, provider, "users"); err != nil {
		return err
	}
	return nil
}
