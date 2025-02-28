package sessionRepository_test

import (
	"context"
	"encoding/json"
	"testing"

	sessionService "github.com/hengadev/leviosa/internal/domain/session"
	miniredis "github.com/hengadev/leviosa/internal/repository/redis"
	sessionRepository "github.com/hengadev/leviosa/internal/repository/redis/session"
	"github.com/hengadev/leviosa/tests/utils/factories"

	assert "github.com/hengadev/test-assert"
)

func TestCreateSession(t *testing.T) {
	session := factories.NewBasicSession(nil)
	tests := []struct {
		name        string
		session     *sessionService.Session
		initMap     miniredis.InitMap[*sessionService.Values]
		expectedErr error
	}{
		{
			name:        "nominal case",
			session:     session,
			initMap:     nil,
			expectedErr: nil,
		},
		{
			name:    "session already exists",
			session: session,
			initMap: map[string]*sessionService.Values{
				session.ID: session.Values(),
			},
			expectedErr: nil,
		},
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
