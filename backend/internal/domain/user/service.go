package userService

import (
	"github.com/hengadev/leviosa/internal/domain/user/security"
	"github.com/hengadev/leviosa/pkg/config"
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
