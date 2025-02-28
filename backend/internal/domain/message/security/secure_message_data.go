package security

import (
	"github.com/hengadev/leviosa/pkg/config"
)

type SecureMessageData struct {
	config *config.SecurityConfig
}

func NewSecureMessageData(config *config.SecurityConfig) *SecureMessageData {
	return &SecureMessageData{config: config}
}
