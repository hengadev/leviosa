package security

import (
	"github.com/GaryHY/leviosa/pkg/config"
)

type SecureEventData struct {
	config *config.SecurityConfig
}

func NewSecureEventData(config *config.SecurityConfig) *SecureEventData {
	return &SecureEventData{config: config}
}
