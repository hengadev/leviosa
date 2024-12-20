package sessionRepository_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/repository/redis"
	"github.com/GaryHY/event-reservation-app/pkg/testutil"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestCreateSession(t *testing.T) {
	tests := []struct {
		session *sessionService.Session
		wantErr bool
		init    miniredis.InitMap[*sessionService.Values]
		name    string
	}{
		{session: nil, wantErr: true, init: nil, name: "nil session"},
		{session: &sessionService.Session{}, wantErr: true, init: nil, name: "empty database"},
		{session: &testutil.BaseSession, wantErr: false, init: nil, name: "nominal case"},
		{session: &testutil.BaseSession, wantErr: false, init: testutil.InitSession, name: "session already exists"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, err := newTestRepository(t, ctx, tt.init)
			if err != nil {
				t.Errorf("setup repository: %s", err)
			}
			sessionEncoded, _ := json.Marshal(tt.session)
			err = repo.CreateSession(ctx, tt.session.ID, sessionEncoded)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
