package eventService_test

import (
	"testing"
	"time"

	"github.com/GaryHY/leviosa/internal/domain/event"
	"github.com/GaryHY/leviosa/internal/domain/event/models"
	"github.com/GaryHY/leviosa/pkg/errsx"
	"github.com/stretchr/testify/assert"
)

func TestParseBeginAt(t *testing.T) {
	tests := []struct {
		name          string
		event         *models.Event
		expectedDay   int
		expectedMonth int
		expectedYear  int
		expectedErrs  errsx.Map
	}{
		{
			name: "Valid BeginAt",
			event: &models.Event{
				BeginAt: time.Date(2023, 10, 5, 0, 0, 0, 0, time.UTC),
			},
			expectedDay:   5,
			expectedMonth: 10,
			expectedYear:  2023,
			expectedErrs:  nil,
		},
		{
			name: "Invalid BeginAt",
			event: &models.Event{
				BeginAt: time.Time{},
			},
			expectedDay:   0,
			expectedMonth: 0,
			expectedYear:  0,
			expectedErrs:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			day, month, year, errs := eventService.ParseBeginAt(tt.event)
			assert.Equal(t, tt.expectedDay, day)
			assert.Equal(t, tt.expectedMonth, month)
			assert.Equal(t, tt.expectedYear, year)
			assert.Equal(t, tt.expectedErrs, errs)
		})
	}
}
