package security

import (
	"github.com/GaryHY/leviosa/pkg/config"
)

type SecureUserData struct {
	config *config.SecurityConfig
}

func NewSecureUserData(config *config.SecurityConfig) *SecureUserData {
	return &SecureUserData{config: config}
}
