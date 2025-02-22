package messageService

import (
	"github.com/GaryHY/leviosa/internal/domain/message/security"
	"github.com/GaryHY/leviosa/pkg/config"
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
