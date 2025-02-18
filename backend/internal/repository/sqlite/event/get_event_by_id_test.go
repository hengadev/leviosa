package eventRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain/event/models"
	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/event"
	"github.com/GaryHY/leviosa/tests/utils"

	"github.com/GaryHY/test-assert"
)

func TestGetEventByID(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		id            string
		expectedEvent *models.Event
		expectedErr   error
		version       int64
		name          string
	}{
		{id: test.GenerateRandomString(12), expectedEvent: nil, expectedErr: nil, version: 20240820013106, name: "no event in database"},
		{id: "ea1d74e2-1612-47ec-aee9-c6a46b65640f", expectedEvent: baseEvent, expectedErr: nil, version: 20240820023513, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, eventRepository.New)
			defer teardown()
			event, err := repo.GetEventByID(ctx, tt.id)
			assert.EqualError(t, err, tt.expectedErr)
			assert.ReflectEqual(t, event, tt.expectedEvent)
		})
	}
}
