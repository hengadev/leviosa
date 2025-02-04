package eventRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/event"
	test "github.com/GaryHY/leviosa/tests"

	"github.com/GaryHY/test-assert"
)

func TestRemoveEvent(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")

	randEventID := test.GenerateRandomString(16)

	tests := []struct {
		eventID string
		wantErr bool
		version int64
		name    string
	}{
		{eventID: baseEventWithPriceID.ID, wantErr: true, version: 20240820013106, name: "no event in database"},
		{eventID: randEventID, wantErr: true, version: 20240820023513, name: "ID do not exist in database"},
		{eventID: baseEventWithPriceID.ID, wantErr: false, version: 20240820023513, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, eventRepository.New)
			defer teardown()
			err := repo.RemoveEvent(ctx, tt.eventID)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
