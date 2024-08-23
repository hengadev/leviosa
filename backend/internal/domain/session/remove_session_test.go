package sessionService_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/tests"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestRemoveSession(t *testing.T) {
	tests := []struct {
		sessionID string
		initMap   KVMap
		wantErr   bool
		name      string
	}{
		{sessionID: baseSession.ID, initMap: nil, wantErr: true, name: "empty repository"},
		{sessionID: test.GenerateRandomString(12), initMap: initMap, wantErr: true, name: "id not in database"},
		{sessionID: baseSession.ID, initMap: initMap, wantErr: false, name: "nominal case"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo := NewStubSessionRepository(ctx, tt.initMap)
			service := sessionService.New(repo)
			err := service.RemoveSession(ctx, tt.sessionID)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
