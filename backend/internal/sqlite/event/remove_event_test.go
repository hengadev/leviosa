package eventRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/sqlite"
	"github.com/GaryHY/event-reservation-app/internal/sqlite/event"
	test "github.com/GaryHY/event-reservation-app/tests"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestRemoveEvent(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")

	randEventID := test.GenerateRandomString(16)

	tests := []struct {
		eventID         string
		expectedEventID string
		wantErr         bool
		version         int64
		name            string
	}{
		{eventID: baseEventWithPriceID.ID, expectedEventID: "", wantErr: true, version: 20240820013106, name: "no event in database"},
		{eventID: randEventID, expectedEventID: "", wantErr: true, version: 20240820023513, name: "ID do not exist in database"},
		{eventID: baseEventWithPriceID.ID, expectedEventID: baseEventWithPriceID.ID, wantErr: false, version: 20240820023513, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, eventRepository.New)
			defer teardown()
			eventID, err := repo.RemoveEvent(ctx, tt.eventID)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, eventID, tt.expectedEventID)
		})
	}
}
