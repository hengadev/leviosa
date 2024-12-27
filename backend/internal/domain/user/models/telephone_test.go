package models_test

import (
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestValidateTelephone(t *testing.T) {
	// TODO: cases
	// - container letters
	// - not 10 numbers
	// - do not start by 0
	// - has spaces
	tests := []struct {
		telephone string
		wantErr   bool
		name      string
	}{
		{telephone: "012345678A", wantErr: true, name: "Contain letters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := models.ValidateTelephone(tt.telephone)
			assert.Equal(t, got != nil, tt.wantErr)
		})
	}
}
