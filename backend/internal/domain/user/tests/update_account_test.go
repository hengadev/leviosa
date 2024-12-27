package models_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/pkg/config"
	"github.com/GaryHY/event-reservation-app/tests/assert"

	"github.com/google/uuid"
)

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		user          *models.User
		mockRepo      func() *MockRepo
		expectedError error
	}{
		{
			name:   "invalid user ID",
			userID: "",
			user:   nil,
			mockRepo: func() *MockRepo {
				return &MockRepo{}
			},
			expectedError: domain.ErrInvalidValue,
		},
		{
			name:   "invalid user",
			userID: uuid.NewString(),
			user:   getInvalidUser(),
			mockRepo: func() *MockRepo {
				return &MockRepo{}
			},
			expectedError: domain.ErrInvalidValue,
		},
		{
			name:   "internal errror",
			userID: uuid.NewString(),
			user:   getValidUser(),
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyAccountFunc: func(ctx context.Context, user *models.User, whereMap map[string]any, prohibitedFields ...string) error {
						return rp.ErrInternal
					},
				}
			},
			expectedError: rp.ErrInternal,
		},
		{
			name:   "context errror",
			userID: uuid.NewString(),
			user:   getValidUser(),
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyAccountFunc: func(ctx context.Context, user *models.User, whereMap map[string]any, prohibitedFields ...string) error {
						return rp.ErrContext
					},
				}
			},
			expectedError: rp.ErrContext,
		},
		{
			name:   "database errror",
			userID: uuid.NewString(),
			user:   getValidUser(),
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyAccountFunc: func(ctx context.Context, user *models.User, whereMap map[string]any, prohibitedFields ...string) error {
						return rp.ErrDatabase
					},
				}
			},
			expectedError: domain.ErrQueryFailed,
		},
		{
			name:   "not updated errror",
			userID: uuid.NewString(),
			user:   getValidUser(),
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyAccountFunc: func(ctx context.Context, user *models.User, whereMap map[string]any, prohibitedFields ...string) error {
						return rp.ErrNotUpdated
					},
				}
			},
			expectedError: domain.ErrNotUpdated,
		},
		{
			name:   "unexpected type errror",
			userID: uuid.NewString(),
			user:   getValidUser(),
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyAccountFunc: func(ctx context.Context, user *models.User, whereMap map[string]any, prohibitedFields ...string) error {
						return errors.New("unexpected type error")
					},
				}
			},
			expectedError: domain.ErrUnexpectedType,
		},
		{
			name:   "successul case",
			userID: uuid.NewString(),
			user:   getValidUser(),
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyAccountFunc: func(ctx context.Context, user *models.User, whereMap map[string]any, prohibitedFields ...string) error {
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
			config := &config.SecurityConfig{}
			service := userService.New(repo, config)
			err := service.UpdateAccount(
				context.Background(),
				tt.user,
				tt.userID,
			)
			assert.EqualError(t, err, tt.expectedError)
		})
	}
}

func getInvalidUser() *models.User {
	return &models.User{
		BirthDate: "",
		LastName:  "DOE",
		FirstName: "John",
		Gender:    "M",
		Telephone: "",
	}
}

func getValidUser() *models.User {
	return &models.User{
		BirthDate: "11-07-1998",
		LastName:  "DOE",
		FirstName: "John",
		Gender:    "M",
		Telephone: "0102345678",
	}
}
