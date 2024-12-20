package userService_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/tests/assert"

	"github.com/google/uuid"
)

func TestGetUserSessionData(t *testing.T) {
	email := "john.doe@hotmail.com"
	userID := uuid.NewString()
	tests := []struct {
		name          string
		email         string
		mockRepo      func() *MockRepo
		expectedError error
		expectedID    string
		expectedRole  userService.Role
	}{
		{
			name:  "invalid email",
			email: "",
			mockRepo: func() *MockRepo {
				return &MockRepo{}
			},
			expectedError: domain.ErrInvalidValue,
			expectedID:    "",
			expectedRole:  userService.UNKNOWN,
		},
		{
			name:  "database error",
			email: email,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					GetUserSessionDataFunc: func(ctx context.Context, email string) (string, userService.Role, error) {
						return "", userService.UNKNOWN, rp.ErrDatabase
					},
				}
			},
			expectedError: domain.ErrQueryFailed,
			expectedID:    "",
			expectedRole:  userService.UNKNOWN,
		},
		{
			name:  "context error",
			email: email,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					GetUserSessionDataFunc: func(ctx context.Context, email string) (string, userService.Role, error) {
						return "", userService.UNKNOWN, rp.ErrContext
					},
				}
			},
			expectedError: rp.ErrContext,
			expectedID:    "",
			expectedRole:  userService.UNKNOWN,
		},
		{
			name:  "unexpected error",
			email: email,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					GetUserSessionDataFunc: func(ctx context.Context, email string) (string, userService.Role, error) {
						return "", userService.UNKNOWN, errors.New("unexpected error")
					},
				}
			},
			expectedError: domain.ErrUnexpectedType,
			expectedID:    "",
			expectedRole:  userService.UNKNOWN,
		},
		{
			name:  "successful case",
			email: email,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					GetUserSessionDataFunc: func(ctx context.Context, email string) (string, userService.Role, error) {
						return userID, userService.BASIC, nil
					},
				}
			},
			expectedError: nil,
			expectedID:    userID,
			expectedRole:  userService.BASIC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := tt.mockRepo()
			service := userService.New(repo)
			id, role, err := service.GetUserSessionData(context.Background(), tt.email)
			assert.EqualError(t, err, tt.expectedError)
			assert.Equal(t, id, tt.expectedID)
			assert.Equal(t, role, tt.expectedRole)
		})
	}
}
