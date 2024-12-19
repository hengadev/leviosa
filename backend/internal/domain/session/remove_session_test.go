package sessionService_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/tests/assert"

	"github.com/google/uuid"
)

func TestRemoveSession(t *testing.T) {
	sessionID := uuid.NewString()
	tests := []struct {
		name          string
		sessionID     string
		mockRepo      func() *MockRepo
		expectedError error
	}{
		{
			name:      "invalid session ID [not UUID]",
			sessionID: "",
			mockRepo: func() *MockRepo {
				return &MockRepo{}
			},
			expectedError: domain.ErrInvalidValue,
		},
		{
			name:      "database error",
			sessionID: sessionID,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					RemoveSessionFunc: func(ctx context.Context, sessionID string) error {
						return rp.ErrDatabase
					},
				}
			},
			expectedError: domain.ErrQueryFailed,
		},
		{
			name:      "context error",
			sessionID: sessionID,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					RemoveSessionFunc: func(ctx context.Context, sessionID string) error {
						return rp.ErrContext
					},
				}
			},
			expectedError: rp.ErrContext,
		},
		{
			name:      "unexpected error",
			sessionID: sessionID,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					RemoveSessionFunc: func(ctx context.Context, sessionID string) error {
						return errors.New("unexpected error")
					},
				}
			},
			expectedError: domain.ErrUnexpectedType,
		},
		{
			name:      "key does not exist",
			sessionID: sessionID,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					RemoveSessionFunc: func(ctx context.Context, sessionID string) error {
						return rp.ErrNotFound
					},
				}
			},
			expectedError: domain.ErrNotFound,
		},
		{
			name:      "successful case",
			sessionID: sessionID,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					RemoveSessionFunc: func(ctx context.Context, sessionID string) error {
						return nil
					},
				}
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := tt.mockRepo()
			service := sessionService.New(repo)
			err := service.RemoveSession(context.Background(), tt.sessionID)
			assert.EqualError(t, err, tt.expectedError)
		})
	}
}
