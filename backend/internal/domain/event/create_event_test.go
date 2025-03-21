package eventService_test

import (
	"context"
	"testing"
	"time"

	"github.com/hengadev/leviosa/internal/domain"
	"github.com/hengadev/leviosa/internal/domain/event"
	"github.com/hengadev/leviosa/internal/domain/event/models"
	rp "github.com/hengadev/leviosa/internal/repository"
	"github.com/hengadev/leviosa/tests/utils"
	"github.com/hengadev/leviosa/tests/utils/factories"

	assert "github.com/hengadev/test-assert"
)

func TestCreateEvent(t *testing.T) {
	conf := test.PrepareEncryptionConfig()
	var zeroTime time.Time
	tests := []struct {
		name          string
		mockRepo      func() *MockRepo
		event         *models.Event
		expectEventID bool
		expectedErr   error
	}{
		{
			name:     "event with BeginAt zero value of time.Time",
			mockRepo: func() *MockRepo { return &MockRepo{} },
			event: factories.NewBasicEvent(map[string]any{
				"BeginAt": zeroTime,
			}),
			expectEventID: false,
			expectedErr:   domain.ErrInvalidValue,
		},
		{
			name: "date is not available because already in database",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					IsDateAvailableFunc: func(ctx context.Context, day, month, year int) error {
						return rp.ErrValidation
					},
				}
			},
			event: factories.NewBasicEvent(map[string]any{
				"BeginAt": time.Now(),
			}),
			expectEventID: false,
			expectedErr:   domain.ErrNotCreated,
		},
		{
			name: "date is not available because context error in database",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					IsDateAvailableFunc: func(ctx context.Context, day, month, year int) error {
						return rp.ErrContext
					},
				}
			},
			event: factories.NewBasicEvent(map[string]any{
				"BeginAt": time.Now(),
			}),
			expectEventID: false,
			expectedErr:   rp.ErrContext,
		},
		{
			name: "date is not available because of a database error",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					IsDateAvailableFunc: func(ctx context.Context, day, month, year int) error {
						return rp.ErrDatabase
					},
				}
			},
			event: factories.NewBasicEvent(map[string]any{
				"BeginAt": time.Now(),
			}),
			expectEventID: false,
			expectedErr:   domain.ErrQueryFailed,
		},
		{
			name: "event not added du to context error",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					AddEventFunc: func(ctx context.Context, event *models.Event) (string, error) {
						return "", rp.ErrContext
					},
				}
			},
			event: factories.NewBasicEvent(map[string]any{
				"BeginAt": time.Now(),
			}),
			expectEventID: false,
			expectedErr:   rp.ErrContext,
		},
		{
			// TODO: that thing is not working, why ?
			name: "event not added du to database error",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					AddEventFunc: func(ctx context.Context, event *models.Event) (string, error) {
						return "", rp.ErrDatabase
					},
				}
			},
			event: factories.NewBasicEvent(map[string]any{
				"BeginAt": time.Now(),
			}),
			expectEventID: false,
			expectedErr:   domain.ErrQueryFailed,
		},
		{
			name: "event not added due to database constraints",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					AddEventFunc: func(ctx context.Context, event *models.Event) (string, error) {
						return "", rp.ErrNotCreated
					},
				}
			},
			event: factories.NewBasicEvent(map[string]any{
				"BeginAt": time.Now(),
			}),
			expectEventID: false,
			expectedErr:   domain.ErrNotCreated,
		},
		{
			name: "nominal case",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					AddEventFunc: func(ctx context.Context, event *models.Event) (string, error) {
						return event.ID, nil
					},
				}
			},
			event: factories.NewBasicEvent(map[string]any{
				"BeginAt": time.Now(),
			}),
			expectEventID: true,
			expectedErr:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			service := eventService.New(tt.mockRepo(), conf)
			eventID, err := service.CreateEvent(ctx, tt.event)
			assert.EqualError(t, err, tt.expectedErr)
			assert.Equal(t, eventID != "", tt.expectEventID)
		})
	}
}
