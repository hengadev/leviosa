package sessionRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/repository/redis"
	"github.com/GaryHY/event-reservation-app/pkg/testutil"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestFindSessionByID(t *testing.T) {
	tests := []struct {
		id              string
		wantErr         bool
		init            miniredis.InitMap[*sessionService.Values]
		expectedSession *sessionService.Session
		name            string
	}{
		{id: testutil.BaseSession.ID, wantErr: true, init: nil, expectedSession: nil, name: "empty database"},
		{id: testutil.RandomSessionID, wantErr: true, init: testutil.InitSession, expectedSession: nil, name: "ID is not in database"},
		{id: testutil.BaseSession.ID, wantErr: false, init: testutil.InitSession, expectedSession: &testutil.BaseSession, name: "nominal case"},
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
