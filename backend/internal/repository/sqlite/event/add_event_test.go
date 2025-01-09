package eventRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain/event"
	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/event"
	"github.com/GaryHY/leviosa/tests/assert"
)

func TestAddEvent(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		event   *eventService.Event
		wantErr bool
		version int64
		name    string
	}{
		{event: &eventService.Event{}, wantErr: true, version: 20240820013106, name: "no event IDa specified"},
		{event: baseEvent, wantErr: true, version: 20240820013106, name: "No price id specified"},
		{event: baseEventWithPriceID, wantErr: false, version: 20240820013106, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, eventRepository.New)
			defer teardown()
			eventID, err := repo.AddEvent(ctx, tt.event)
			_ = eventID
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
