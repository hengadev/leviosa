package mailService

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/pkg/errsx"
)

func (s *Service) PendingUser(ctx context.Context, user *models.User) errsx.Map {
	var errs errsx.Map
	return errs
}
