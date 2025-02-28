package messageService

import (
	"github.com/hengadev/leviosa/internal/domain/message/security"
	"github.com/hengadev/leviosa/pkg/config"
)

type Service struct {
	repo ReadWriter
	*security.SecureMessageData
}

func New(repo ReadWriter, conf *config.SecurityConfig) *Service {
	return &Service{
		repo,
		security.NewSecureMessageData(conf)}
}
