package sessionService_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	test "github.com/GaryHY/event-reservation-app/tests"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestGetSession(t *testing.T) {
	// TEST:
	tests := []struct {
		sessionID       string
		initMap         KVMap
		expectedSession *sessionService.Session
		wantErr         bool
		name            string
	}{
		{sessionID: baseSession.ID, initMap: nil, expectedSession: nil, wantErr: true, name: "no session in database"},
		{sessionID: test.GenerateRandomString(12), initMap: initMap, expectedSession: nil, wantErr: true, name: "id not in database"},
		{sessionID: baseSession.ID, initMap: initMap, expectedSession: baseSession, wantErr: false, name: "nominal case"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo := NewStubSessionRepository(ctx, tt.initMap)
			service := sessionService.New(repo)
			sess, err := service.GetSession(ctx, tt.sessionID)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.ReflectEqual(t, sess, tt.expectedSession)

		})
	}
}
