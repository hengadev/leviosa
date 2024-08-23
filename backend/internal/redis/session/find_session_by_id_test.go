package sessionRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	miniredis "github.com/GaryHY/event-reservation-app/internal/redis"
	"github.com/GaryHY/event-reservation-app/tests"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestFindSessionByID(t *testing.T) {
	tests := []struct {
		id              string
		wantErr         bool
		init            miniredis.InitMap
		expectedSession *sessionService.Session
		name            string
	}{
		{id: baseSession.ID, wantErr: true, init: nil, expectedSession: nil, name: "empty database"},
		{id: test.GenerateRandomString(12), wantErr: true, init: initSession, expectedSession: nil, name: "ID is not in database"},
		{id: baseSession.ID, wantErr: false, init: initSession, expectedSession: &baseSession, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, err := newTestRepository(t, ctx, tt.init)
			if err != nil {
				t.Errorf("setup repository: %s", err)
			}
			res, err := repo.FindSessionByID(ctx, tt.id)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.ReflectEqual(t, res, tt.expectedSession)
		})
	}
}
