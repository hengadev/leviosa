package eventRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/repository/sqlite"
	"github.com/GaryHY/leviosa/internal/repository/sqlite/event"
	"github.com/GaryHY/leviosa/tests"

	"github.com/GaryHY/test-assert"
)

func TestGetPriceID(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	tests := []struct {
		eventID         string
		expectedPriceID string
		wantErr         bool
		version         int64
		name            string
	}{
		{eventID: baseEventWithPriceID.ID, expectedPriceID: "", wantErr: true, version: 20240820013106, name: "No event in database"},
		{eventID: test.GenerateRandomString(12), expectedPriceID: "", wantErr: true, version: 20240820023513, name: "ID not in database"},
		{eventID: baseEventWithPriceID.ID, expectedPriceID: baseEventWithPriceID.PriceID, wantErr: false, version: 20240820023513, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, teardown := sqlite.SetupRepository(t, ctx, tt.version, eventRepository.New)
			defer teardown()
			priceID, err := repo.GetPriceID(ctx, tt.eventID)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.ReflectEqual(t, priceID, tt.expectedPriceID)
		})
	}
}
