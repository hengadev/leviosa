package eventService_test

import (
	"errors"
	"testing"
	"time"

	"github.com/hengadev/leviosa/internal/domain/event"
	"github.com/hengadev/leviosa/internal/domain/event/models"
	"github.com/hengadev/leviosa/tests/utils/factories"
	"github.com/stretchr/testify/assert"
)

func TestParseBeginAt(t *testing.T) {
	var zeroTime time.Time
	now := time.Now()
	tests := []struct {
		name          string
		event         *models.Event
		expectedDay   int
		expectedMonth int
		expectedYear  int
		expectedErr   error
	}{
		{
			name: "Valid BeginAt",
			event: factories.NewBasicEvent(map[string]any{
				"BeginAt": now,
			}),
			expectedDay:   now.Day(),
			expectedMonth: int(now.Month()),
			expectedYear:  now.Year(),
			expectedErr:   nil,
		},
		{
			name: "Invalid BeginAt",
			event: &models.Event{
				BeginAt: zeroTime,
			},
			expectedDay:   0,
			expectedMonth: 0,
			expectedYear:  0,
			expectedErr:   errors.New("BeginAt is zero"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			day, month, year, errs := eventService.ParseBeginAt(tt.event)
			assert.Equal(t, tt.expectedDay, day)
			assert.Equal(t, tt.expectedMonth, month)
			assert.Equal(t, tt.expectedYear, year)
			assert.Equal(t, tt.expectedErr, errs)
		})
	}
}
