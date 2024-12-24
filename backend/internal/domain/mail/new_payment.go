package mailService

import (
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/pkg/errsx"
)

func (s *Service) NewPayment(user *models.User, eventTime string) errsx.Map {
	var errs errsx.Map
	return errs
}
