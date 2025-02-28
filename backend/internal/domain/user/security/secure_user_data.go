package security

import (
	"github.com/hengadev/leviosa/pkg/config"
)

type SecureUserData struct {
	config *config.SecurityConfig
}

func NewSecureUserData(config *config.SecurityConfig) *SecureUserData {
	return &SecureUserData{config: config}
}
