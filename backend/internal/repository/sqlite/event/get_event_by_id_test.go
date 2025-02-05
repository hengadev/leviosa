package eventRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain/event"
	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/event"
	"github.com/GaryHY/leviosa/tests/utils"

	"github.com/GaryHY/test-assert"
)

func TestGetEventByID(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		id            string
		expectedEvent *eventService.Event
		wantErr       bool
		version       int64
		name          string
	}{
		{id: test.GenerateRandomString(12), expectedEvent: nil, wantErr: true, version: 20240820013106, name: "no event in database"},
		{id: "ea1d74e2-1612-47ec-aee9-c6a46b65640f", expectedEvent: baseEvent, wantErr: false, version: 20240820023513, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, eventRepository.New)
			defer teardown()
			event, err := repo.GetEventByID(ctx, tt.id)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.ReflectEqual(t, event, tt.expectedEvent)
		})
	}
}
