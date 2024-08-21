package sessionRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	miniredis "github.com/GaryHY/event-reservation-app/internal/redis"

	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestCreateSession(t *testing.T) {
	tests := []struct {
		session *session.Session
		wantErr bool
		init    miniredis.InitMap
		name    string
	}{
		{session: nil, wantErr: true, init: nil, name: "nil session"},
		{session: &session.Session{}, wantErr: true, init: nil, name: "empty database"},
		{session: &baseSession, wantErr: false, init: nil, name: "nominal case"},
		{session: &baseSession, wantErr: false, init: initSession, name: "session already exists"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, err := newTestRepository(t, ctx, tt.init)
			if err != nil {
				t.Errorf("setup repository: %s", err)
			}
			err = repo.CreateSession(ctx, tt.session)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
