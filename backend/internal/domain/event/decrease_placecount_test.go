package eventService_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain"
	eventService "github.com/GaryHY/leviosa/internal/domain/event"
	rp "github.com/GaryHY/leviosa/internal/repository"
	test "github.com/GaryHY/leviosa/tests/utils"
	"github.com/GaryHY/leviosa/tests/utils/factories"
	assert "github.com/GaryHY/test-assert"
)

func TestDecreasePlacecount(t *testing.T) {
	conf := test.PrepareEncryptionConfig()
	event := factories.NewBasicEvent(nil)

	tests := []struct {
		name        string
		mockRepo    func() *MockRepo
		eventID     string
		expectedErr error
	}{
		{
			name: "context error",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					DecreaseFreePlaceFunc: func(ctx context.Context, eventID string) error {
						return rp.ErrContext
					},
				}
			},
			eventID:     event.ID,
			expectedErr: rp.ErrContext,
		},
		{
			name: "database error",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					DecreaseFreePlaceFunc: func(ctx context.Context, eventID string) error {
						return rp.ErrDatabase
					},
				}
			},
			eventID:     event.ID,
			expectedErr: domain.ErrQueryFailed,
		},
		{
			name: "not updated error",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					DecreaseFreePlaceFunc: func(ctx context.Context, eventID string) error {
						return domain.ErrNotUpdated
					},
				}
			},
			eventID:     event.ID,
			expectedErr: domain.ErrNotUpdated,
		},
		{
			name: "random error",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					DecreaseFreePlaceFunc: func(ctx context.Context, eventID string) error {
						return errors.New("some random error")
					},
				}
			},
			eventID:     event.ID,
			expectedErr: domain.ErrUnexpectedType,
		},
		{
			name: "nominal case",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					DecreaseFreePlaceFunc: func(ctx context.Context, eventID string) error {
						return nil
					},
				}
			},
			eventID:     event.ID,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			service := eventService.New(tt.mockRepo(), conf)
			err := service.DecreasePlacecount(ctx, tt.eventID)
			assert.EqualError(t, err, tt.expectedErr)
		})
	}
}
