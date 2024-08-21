package sessionRepository_test

import (
	"context"
	"testing"

	miniredis "github.com/GaryHY/event-reservation-app/internal/redis"
	test "github.com/GaryHY/event-reservation-app/tests"

	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestRemoveSession(t *testing.T) {
	tests := []struct {
		id      string
		wantErr bool
		init    miniredis.InitMap
		name    string
	}{
		{id: baseSession.ID, wantErr: true, init: nil, name: "empty database"},
		{id: test.GenerateRandomString(12), wantErr: true, init: initSession, name: "id not in the database"},
		{id: baseSession.ID, wantErr: false, init: initSession, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, err := newTestRepository(t, ctx, tt.init)
			if err != nil {
				t.Errorf("setup repository: %s", err)
			}
			err = repo.RemoveSession(ctx, tt.id)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
