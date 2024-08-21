package eventRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/sqlite"
	"github.com/GaryHY/event-reservation-app/internal/sqlite/event"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestGetAllEvents(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		expectedEvents []*event.Event
		wantErr        bool
		version        int64
		name           string
	}{
		{expectedEvents: nil, wantErr: false, version: 20240820013106, name: "no event in database"},
		{expectedEvents: []*event.Event{baseEvent, baseEvent1, baseEvent2}, wantErr: false, version: 20240820103230, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, eventRepository.New)
			defer teardown()
			events, err := repo.GetAllEvents(ctx)
			assert.Equal(t, err != nil, tt.wantErr)
			for i, event := range events {
				assert.ReflectEqual(t, event, tt.expectedEvents[i])
			}
		})
	}
}
