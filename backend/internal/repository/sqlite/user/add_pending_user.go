package userRepository

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
)

// AddUser adds a user to the 'pending_users' table, considering the specified provider type.
// If accounts from other providers exist, it ensures proper account linking.
func (u *Repository) AddPendingUser(ctx context.Context, user *models.User, provider models.ProviderType) error {
	if err := u.addGenericUser(ctx, user, provider, "pending_users"); err != nil {
		return err
	}
	return nil
}
