package eventRepository_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/sqlite/event"
	testdb "github.com/GaryHY/event-reservation-app/pkg/sqliteutil/testdatabase"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

// NOTE: I should get the past event for those the user actually registered for

func TestGetEventForUser(t *testing.T) {
	t.Setenv("TEST_MIGRATION_PATH", "../migrations/tests")

	tests := []struct {
		event           *event.Event
		expectedEventID string
		wantErr         bool
		version         int64
		name            string
	}{
		{},
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
			eventUser, err := repo.AddEvent(ctx, tt.event)
			assert.Equal(t, err != nil, tt.wantErr)
			_ = eventUser
		})
	}
}
