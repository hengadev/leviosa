package eventRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/GaryHY/leviosa/internal/domain/event"
	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/event"

	"github.com/GaryHY/test-assert"
)

func TestModifyEvent(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")

	whereMap := map[string]any{"id": baseEventWithPriceID.ID}

	changes := map[string]any{"Location": "Changed location", "PriceID": "a new price id"}
	modifiedEventProhibitedField, err := domain.CreateWithZeroFieldModifiedObject(*baseEventWithPriceID, changes)
	if err != nil {
		t.Error("Failed to create object with modified field")
	}

	changes2 := map[string]any{"Location": "Changed location"}
	modifiedEvent, err := domain.CreateWithZeroFieldModifiedObject(*baseEventWithPriceID, changes2)
	if err != nil {
		t.Error("Failed to create object with modified field")
	}

	tests := []struct {
		eventModified *eventService.Event
		wantErr       bool
		version       int64
		name          string
	}{
		{eventModified: nil, wantErr: true, version: 20240820013106, name: "nil event"},
		{eventModified: baseEventWithPriceID, wantErr: true, version: 20240820023513, name: "event with prohibited fields for modification"},
		{eventModified: modifiedEventProhibitedField, wantErr: true, version: 20240820023513, name: "make changes to prohibited field"},
		{eventModified: modifiedEvent, wantErr: false, version: 20240820023513, name: "nominal case with valid updatable event"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, eventRepository.New)
			defer teardown()
			err := repo.ModifyEvent(
				ctx, tt.eventModified,
				whereMap,
				"ID",
				"PriceID",
			)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
