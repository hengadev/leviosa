package eventRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain/event/models"
	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/event"

	"github.com/GaryHY/test-assert"
)

func TestGetAllEvents(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		expectedEvents []*models.Event
		wantErr        bool
		version        int64
		name           string
	}{
		{expectedEvents: nil, wantErr: false, version: 20240820013106, name: "no event in database"},
		{expectedEvents: []*models.Event{baseEvent, baseEvent1, baseEvent2}, wantErr: false, version: 20240820103230, name: "nominal case"},
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
