package eventRepository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/sqlite/event"
	testdb "github.com/GaryHY/event-reservation-app/pkg/sqliteutil/testdatabase"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestAddEvent(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")
	fmt.Printf("the event id of smth empty %q\n", event.Event{}.ID)
	tests := []struct {
		event           *event.Event
		expectedEventID string
		wantErr         bool
		version         int64
		name            string
	}{
		{event: &event.Event{}, expectedEventID: "", wantErr: true, version: 20240820013106, name: "no event IDa specified"},
		{event: baseEvent, expectedEventID: "", wantErr: true, version: 20240820013106, name: "No price id specified"},
		{event: baseEventWithPriceID, expectedEventID: baseEvent.ID, wantErr: false, version: 20240820013106, name: "nominal case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, err := testdb.SetupRepo(ctx, tt.version, eventRepository.New)
			defer testdb.Teardown(repo.DB)
			if err != nil {
				t.Errorf("setup event domain stub repository: %s", err)
			}
			eventID, err := repo.AddEvent(ctx, tt.event)
			fmt.Printf("the eventID is: %q\n", eventID)
			fmt.Println("the error is:", err)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, eventID, tt.expectedEventID)
		})
	}
}
