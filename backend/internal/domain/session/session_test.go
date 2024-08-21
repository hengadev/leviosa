package session_test

import (
	"testing"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	test "github.com/GaryHY/event-reservation-app/tests"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestIsZero(t *testing.T) {
	tests := []struct {
		session      session.Session
		name         string
		expectIsZero bool
	}{
		{session: session.Session{}, expectIsZero: true, name: "zero session"},
		{session: session.Session{
			ID:         "",
			UserID:     1,
			Role:       "",
			LoggedInAt: time.Now(),
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(time.Hour),
		}, expectIsZero: true, name: "some field are zero in session"},
		{
			session: session.Session{
				ID:         test.GenerateRandomString(12),
				UserID:     1,
				Role:       "basic",
				LoggedInAt: time.Now(),
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(time.Hour),
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
