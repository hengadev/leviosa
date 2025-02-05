package sessionService_test

import (
	"testing"
	"time"

	"github.com/GaryHY/leviosa/internal/domain/session"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/tests/utils"

	"github.com/GaryHY/test-assert"
)

func TestIsZero(t *testing.T) {
	tests := []struct {
		session      sessionService.Session
		name         string
		expectIsZero bool
	}{
		{session: sessionService.Session{}, expectIsZero: true, name: "zero session"},
		{session: sessionService.Session{
			ID:        "",
			UserID:    "1",
			Role:      models.BASIC,
			ExpiresAt: time.Now().Add(time.Hour),
		}, expectIsZero: true, name: "some field are zero in session"},
		{
			session: sessionService.Session{
				ID:        test.GenerateRandomString(12),
				UserID:    "1",
				Role:      models.BASIC,
				ExpiresAt: time.Now().Add(time.Hour),
			}, expectIsZero: false, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			isZero := tt.session.IsZero()
			assert.Equal(t, isZero, tt.expectIsZero)
		})
	}
}
