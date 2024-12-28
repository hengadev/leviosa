package userService

import (
	"github.com/GaryHY/event-reservation-app/internal/domain/user/security"
	"github.com/GaryHY/event-reservation-app/pkg/config"
)

type Service struct {
	repo ReadWriter
	*security.SecureUserData
}

func New(repo ReadWriter, config *config.SecurityConfig) *Service {
	return &Service{
		repo,
		security.NewSecureUserData(config),
	}
}
