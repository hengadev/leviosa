package userService

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
)

// FindAccountByID return the decrypted user with the specified ID
func (s *Service) FindAccountByID(ctx context.Context, userID string) (*models.User, error) {
	// the encrypted user from the database
	user, err := s.repo.FindAccountByID(ctx, userID)
	if err != nil {
		switch {
		default:
			return nil, domain.NewUnexpectTypeErr(err)
		}
	}
	// - decrypt user
	err = s.DecryptUser(user)
	if err != nil {

	}
	return user, nil
}
