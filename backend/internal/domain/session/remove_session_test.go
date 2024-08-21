package session_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/tests"
)

func TestRemoveSession(t *testing.T) {
	tests := []struct {
		sessionID string
		initMap   KVMap
		wantErr   bool
		name      string
	}{
		{sessionID: baseSession.ID, initMap: nil, wantErr: true, name: "empty repository"},
		{sessionID: test.GenerateRandomString(12), initMap: initSessionValues, wantErr: true, name: "id not in database"},
		{sessionID: baseSession.ID, initMap: initSessionValues, wantErr: false, name: "nominal case"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo := NewStubSessionRepository(ctx, tt.initMap)
			service := session.NewService(repo)
			err := service.RemoveSession(ctx, tt.sessionID)
			test.Assert(t, err != nil, tt.wantErr)
		})
	}
}
