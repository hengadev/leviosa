package services

import (
	"testing"
)

func TestServiceAPIKeyPath(t *testing.T) {
	tests := []struct {
		service  string
		expected string
	}{
		{AuthUser, "secret/data/services/authuser/api-key"},
		{Catalog, "secret/data/services/catalog/api-key"},
		{Settings, "secret/data/services/settings/api-key"},
		{Notification, "secret/data/services/notification/api-key"},
	}

	for _, tt := range tests {
		t.Run(tt.service, func(t *testing.T) {
			result := ServiceAPIKeyPath(tt.service)
			if result != tt.expected {
				t.Errorf("ServiceAPIKeyPath(%s) = %s, expected %s", tt.service, result, tt.expected)
			}
		})
	}
}
