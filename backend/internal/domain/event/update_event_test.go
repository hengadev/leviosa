package eventService_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hengadev/leviosa/internal/domain"
	eventService "github.com/hengadev/leviosa/internal/domain/event"
	"github.com/hengadev/leviosa/internal/domain/event/models"
	rp "github.com/hengadev/leviosa/internal/repository"
	test "github.com/hengadev/leviosa/tests/utils"
	"github.com/hengadev/leviosa/tests/utils/factories"
	assert "github.com/hengadev/test-assert"
)

func TestModifyEvent(t *testing.T) {
	conf := test.PrepareEncryptionConfig()
	event := factories.NewBasicEvent(nil)

	tests := []struct {
		name        string
		mockRepo    func() *MockRepo
		event       *models.Event
		expectedErr error
	}{

		{
			name:        "nil event",
			mockRepo:    func() *MockRepo { return &MockRepo{} },
			event:       nil,
			expectedErr: domain.ErrInvalidValue,
		},
		{
			name: "context error",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyEventFunc: func(ctx context.Context, event *models.Event, whereMap map[string]any) error {
						return rp.ErrContext
					},
				}
			},
			event:       event,
			expectedErr: rp.ErrContext,
		},
		{
			name: "database error",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyEventFunc: func(ctx context.Context, event *models.Event, whereMap map[string]any) error {
						return rp.ErrDatabase
					},
				}
			},
			event:       event,
			expectedErr: domain.ErrQueryFailed,
		},
		{
			name: "internal error",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyEventFunc: func(ctx context.Context, event *models.Event, whereMap map[string]any) error {
						return rp.ErrInternal
					},
				}
			},
			event:       event,
			expectedErr: rp.ErrInternal,
		},
		{
			name: "not updated error",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyEventFunc: func(ctx context.Context, event *models.Event, whereMap map[string]any) error {
						return rp.ErrNotUpdated
					},
				}
			},
			event:       event,
			expectedErr: domain.ErrNotUpdated,
		},
		{
			name: "random error",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyEventFunc: func(ctx context.Context, event *models.Event, whereMap map[string]any) error {
						return errors.New("some random error")
					},
				}
			},
			event:       event,
			expectedErr: domain.ErrUnexpectedType,
		},
		{
			name: "nominal case",
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyEventFunc: func(ctx context.Context, event *models.Event, whereMap map[string]any) error {
						return nil
					},
				}
			},
			event:       event,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			service := eventService.New(tt.mockRepo(), conf)
			err := service.ModifyEvent(ctx, tt.event)
			assert.EqualError(t, err, tt.expectedErr)
		})
	}
}
