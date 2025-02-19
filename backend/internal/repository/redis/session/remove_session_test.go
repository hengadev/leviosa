package sessionRepository_test

import (
	"context"
	"testing"

	sessionService "github.com/GaryHY/leviosa/internal/domain/session"
	rp "github.com/GaryHY/leviosa/internal/repository"
	miniredis "github.com/GaryHY/leviosa/internal/repository/redis"
	sessionRepository "github.com/GaryHY/leviosa/internal/repository/redis/session"
	test "github.com/GaryHY/leviosa/tests/utils"
	"github.com/GaryHY/leviosa/tests/utils/factories"

	assert "github.com/GaryHY/test-assert"
)

func TestRemoveSession(t *testing.T) {
	session := factories.NewBasicSession(nil)
	tests := []struct {
		name        string
		ID          string
		initMap     miniredis.InitMap[*sessionService.Values]
		expectedErr error
	}{
		{
			name:        "empty database",
			ID:          session.ID,
			initMap:     nil,
			expectedErr: rp.ErrNotFound,
		},
		{
			name: "ID not in the database",
			ID:   test.GenerateRandomString(16),
			initMap: map[string]*sessionService.Values{
				session.ID: session.Values(),
			},
			expectedErr: rp.ErrNotFound,
		},
		{
			name: "nominal case",
			ID:   session.ID,
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
			err := repo.RemoveSession(ctx, tt.ID)
			assert.EqualError(t, err, tt.expectedErr)
		})
	}
}
