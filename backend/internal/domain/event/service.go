package eventService

import (
	"github.com/hengadev/leviosa/internal/domain/event/security"
	"github.com/hengadev/leviosa/pkg/config"
)

type Service struct {
	repo ReadWriter
	*security.SecureEventData
}

func New(repo ReadWriter, config *config.SecurityConfig) *Service {
	return &Service{
		repo,
		security.NewSecureEventData(config),
	}
}
