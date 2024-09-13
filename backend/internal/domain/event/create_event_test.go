package eventService_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestCreateEvent(t *testing.T) {
	// TEST:
	// - wrong format event
	// - no event table
	// - event already exists
	// - nominal case

	tests := []struct {
		event   *eventService.Event
		wantErr bool
		name    string
	}{
		// {},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo := NewStubEventRepository()
			service := eventService.New(repo)
			gotUser, gotErr := service.CreateEvent(ctx, tt.event)
			_ = gotUser
			assert.Equal(t, gotErr != nil, tt.wantErr)
		})
	}
}
