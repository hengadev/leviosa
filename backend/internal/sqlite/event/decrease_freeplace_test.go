package eventRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/sqlite"
	"github.com/GaryHY/event-reservation-app/internal/sqlite/event"
	"github.com/GaryHY/event-reservation-app/tests"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestDecreaseFreeplace(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		id      string
		wantErr bool
		version int64
		name    string
	}{
		{id: baseEvent2.ID, wantErr: true, version: 20240820103230, name: "freeplace is equal to 0"},
		{id: test.GenerateRandomString(16), wantErr: true, version: 20240820103230, name: "id not found"},
		{id: baseEvent.ID, wantErr: false, version: 20240820103230, name: "normal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, eventRepository.New)
			defer teardown()
			err := repo.DecreaseFreeplace(ctx, tt.id)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
