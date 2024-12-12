package sessionService_test

import (
	"testing"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	userService "github.com/GaryHY/event-reservation-app/internal/domain/user"
	test "github.com/GaryHY/event-reservation-app/tests"
	"github.com/GaryHY/event-reservation-app/tests/assert"
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
			Role:      userService.BASIC,
			ExpiresAt: time.Now().Add(time.Hour),
		}, expectIsZero: true, name: "some field are zero in session"},
		{
			session: sessionService.Session{
				ID:        test.GenerateRandomString(12),
				UserID:    "1",
				Role:      userService.BASIC,
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
