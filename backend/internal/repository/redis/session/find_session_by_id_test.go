package sessionRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain/session"
	"github.com/GaryHY/leviosa/internal/repository/redis"
	"github.com/GaryHY/leviosa/pkg/testutil"
	"github.com/GaryHY/leviosa/tests/assert"
)

func TestFindSessionByID(t *testing.T) {
	tests := []struct {
		id              string
		wantErr         bool
		init            miniredis.InitMap[*sessionService.Values]
		expectedSession *sessionService.Session
		name            string
	}{
		// NOTE: how to handle mock data when doing testing brother
		{id: testutil.BaseSession.ID, wantErr: true, init: nil, expectedSession: nil, name: "empty database"},
		{id: testutil.RandomSessionID, wantErr: true, init: testutil.InitSession, expectedSession: nil, name: "ID is not in database"},
		{id: testutil.BaseSession.ID, wantErr: false, init: testutil.InitSession, expectedSession: &testutil.BaseSession, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, _ := newTestRepository(t, ctx, tt.init)
			service := sessionService.New(repo)
			session, err := service.GetSession(ctx, tt.id)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.ReflectEqual(t, session, tt.expectedSession)
		})
	}
}
