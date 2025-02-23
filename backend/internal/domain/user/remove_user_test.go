package userService_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/GaryHY/leviosa/internal/domain/user"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/pkg/config"

	"github.com/GaryHY/test-assert"
	"github.com/google/uuid"
)

func TestDeleteUser(t *testing.T) {
	userID := uuid.NewString()
	tests := []struct {
		name          string
		userID        string
		mockRepo      func() *MockRepo
		expectedError error
	}{
		{
			name:   "invalid user ID",
			userID: "",
			mockRepo: func() *MockRepo {
				return &MockRepo{}
			},
			expectedError: domain.ErrInvalidValue,
		},
		{
			name:   "context error",
			userID: userID,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					DeleteUserFunc: func(ctx context.Context, id string) error {
						return rp.ErrContext
					},
				}
			},
			expectedError: rp.ErrContext,
		},
		{
			name:   "database error",
			userID: userID,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					DeleteUserFunc: func(ctx context.Context, id string) error {
						return rp.ErrDatabase
					},
				}
			},
			expectedError: domain.ErrQueryFailed,
		},
		{
			name:   "key does not exist",
			userID: userID,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					DeleteUserFunc: func(ctx context.Context, id string) error {
						return rp.ErrNotDeleted
					},
				}
			},
			expectedError: domain.ErrNotDeleted,
		},
		{
			name:   "successful case",
			userID: userID,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					DeleteUserFunc: func(ctx context.Context, id string) error {
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
			// TODO: find the config for this
			config := &config.SecurityConfig{}
			service := userService.New(repo, config)
			err := service.DeleteUser(context.Background(), tt.userID)
			assert.EqualError(t, err, tt.expectedError)
		})
	}
}
