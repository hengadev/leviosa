package sessionRepository_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain/session"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/internal/repository/redis"
	sessionRepository "github.com/GaryHY/leviosa/internal/repository/redis/session"
	test "github.com/GaryHY/leviosa/tests/utils"
	"github.com/GaryHY/leviosa/tests/utils/factories"

	"github.com/GaryHY/test-assert"
)

func TestFindSessionByID(t *testing.T) {
	session := factories.NewBasicSession(nil)
	values, err := json.Marshal(session.Values())
	_ = values
	if err != nil {
		t.Fatalf("failed to marshal expected session values: %s", err)
	}
	tests := []struct {
		name    string
		id      string
		initMap miniredis.InitMap[*sessionService.Values]
		// expectedValues *sessionService.Values
		expectedValues []byte
		expectedErr    error
	}{
		{
			name:           "empty database",
			id:             session.ID,
			initMap:        nil,
			expectedValues: []byte{},
			expectedErr:    rp.ErrNotFound,
		},
		{
			name: "ID is not in database",
			id:   test.GenerateRandomString(16),
			initMap: map[string]*sessionService.Values{
				session.ID: session.Values(),
			},
			expectedValues: []byte{},
			expectedErr:    rp.ErrNotFound,
		},
		{
			name: "nominal case",
			id:   session.ID,
			initMap: map[string]*sessionService.Values{
				session.ID: session.Values(),
			},
			expectedValues: values,
			expectedErr:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo := miniredis.SetupRepository(t, ctx, sessionRepository.SESSIONPREFIX, tt.initMap, sessionRepository.New)
			encodedSession, err := repo.FindSessionByID(ctx, tt.id)
			assert.EqualError(t, err, tt.expectedErr)
			if !bytes.Equal(encodedSession, tt.expectedValues) {
				t.Errorf("got %v, want %v", encodedSession, tt.expectedValues)
			}
		})
	}
}
