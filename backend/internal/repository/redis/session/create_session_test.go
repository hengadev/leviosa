package sessionRepository_test

import (
	"context"
	"encoding/json"
	"testing"

	sessionService "github.com/GaryHY/leviosa/internal/domain/session"
	miniredis "github.com/GaryHY/leviosa/internal/repository/redis"
	sessionRepository "github.com/GaryHY/leviosa/internal/repository/redis/session"
	"github.com/GaryHY/leviosa/tests/utils/factories"

	assert "github.com/GaryHY/test-assert"
)

func TestCreateSession(t *testing.T) {
	tests := []struct {
		name        string
		session     *sessionService.Session
		initMap     miniredis.InitMap[*sessionService.Values]
		expectedErr error
	}{
		{session: nil, expectedErr: nil, initMap: nil, name: "nil session"},
		{session: &sessionService.Session{}, expectedErr: nil, initMap: nil, name: "empty database"},
		{session: &factories.BaseSession, expectedErr: nil, initMap: nil, name: "nominal case"},
		{session: &factories.BaseSession, expectedErr: nil, initMap: factories.InitSession, name: "session already exists"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo := miniredis.SetupRepository(t, ctx, sessionRepository.SESSIONPREFIX, tt.initMap, sessionRepository.New)
			sessionEncoded, _ := json.Marshal(tt.session)
			err := repo.CreateSession(ctx, tt.session.ID, sessionEncoded)
			assert.EqualError(t, err, tt.expectedErr)
		})
	}
}
