package session_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/tests"
)

func TestCreateSession(t *testing.T) {
	// TODO: cases:
	// - no session : create the session
	// - has valid session
	tests := []struct {
		sess          *session.Session
		sessions      Values
		wantErr       bool
		wantSessionID bool
		name          string
	}{
		{sess: sessionTest, sessions: initSessionValues, wantErr: false, wantSessionID: true, name: "no session already"},
		{sess: sessionTest, sessions: Values{sessionTest.ID: sessionTest.Values()}, wantErr: false, wantSessionID: true, name: "session already exists"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo := NewStubSessionRepository(ctx, tt.sessions)
			service := session.NewService(repo)
			sessionID, err := service.CreateSession(ctx, 1, user.BASIC.String())
			test.Assert(t, sessionID != "", tt.wantSessionID)
			test.Assert(t, err != nil, tt.wantErr)
		})
	}
}
