package sessionRepository_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain/session"
	"github.com/GaryHY/leviosa/internal/repository/redis"
	"github.com/GaryHY/leviosa/pkg/testutil"

	"github.com/GaryHY/test-assert"
)

func TestCreateSession(t *testing.T) {
	tests := []struct {
		name        string
		session     *sessionService.Session
		init        miniredis.InitMap[*sessionService.Values]
		expectedErr error
	}{
		{session: nil, expectedErr: nil, init: nil, name: "nil session"},
		{session: &sessionService.Session{}, expectedErr: nil, init: nil, name: "empty database"},
		{session: &testutil.BaseSession, expectedErr: nil, init: nil, name: "nominal case"},
		{session: &testutil.BaseSession, expectedErr: nil, init: testutil.InitSession, name: "session already exists"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			// TODO: remove the newtnewTestRepository function for something more general about redis, it would be for the other tests.
			repo, err := newTestRepository(t, ctx, tt.init)
			if err != nil {
				t.Errorf("setup repository: %s", err)
			}
			sessionEncoded, _ := json.Marshal(tt.session)
			err = repo.CreateSession(ctx, tt.session.ID, sessionEncoded)
			assert.EqualError(t, nil, tt.expectedErr)
		})
	}
}
