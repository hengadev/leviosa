package userRepository

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
)

func (u *Repository) GetPendingUsers(ctx context.Context) ([]*models.User, error) {
	return nil, nil
}
