package session_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestCreateSession(t *testing.T) {
	// TODO: cases:
	// - no session : create the session
	// - has valid session
	tests := []struct {
		sess          *session.Session
		sessions      KVMap
		wantErr       bool
		wantSessionID bool
		name          string
	}{
		{sess: baseSession, sessions: KVMap{}, wantErr: false, wantSessionID: true, name: "no session already"},
		{sess: baseSession, sessions: KVMap{baseSession.ID: baseSession.Values()}, wantErr: false, wantSessionID: true, name: "session already exists"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo := NewStubSessionRepository(ctx, tt.sessions)
			service := session.NewService(repo)
			sessionID, err := service.CreateSession(ctx, 1, user.BASIC.String())
			assert.Equal(t, sessionID != "", tt.wantSessionID)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
