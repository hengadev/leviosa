package mailService

import (
	"context"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/pkg/errsx"
)

func (s *Service) PendingUser(ctx context.Context, user *models.User) errsx.Map {
	var errs errsx.Map
	return errs
}
