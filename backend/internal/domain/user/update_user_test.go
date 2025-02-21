package userService_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain"
	userService "github.com/GaryHY/leviosa/internal/domain/user"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/tests/utils"
	"github.com/GaryHY/leviosa/tests/utils/factories"

	assert "github.com/GaryHY/test-assert"
)

func TestUpdateUser(t *testing.T) {
	conf := test.PrepareEncryptionConfig()
	user := factories.NewBasicUser(nil)
	tests := []struct {
		name          string
		user          *models.User
		mockRepo      func() *MockRepo
		expectedError error
	}{
		{
			name: "user ID not uuid",
			user: factories.NewBasicUser(map[string]any{
				"ID": test.GenerateRandomString(16),
			}),
			mockRepo:      func() *MockRepo { return &MockRepo{} },
			expectedError: domain.ErrInvalidValue,
		},
		{
			name: "ModifyAccount writing update query errror",
			user: user,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyAccountFunc: func(ctx context.Context, user *models.User, whereMap map[string]any) error {
						return rp.ErrValidation
					},
				}
			},
			expectedError: domain.ErrInvalidValue,
		},
		{
			name: "ModifyAccount context errror",
			user: user,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyAccountFunc: func(ctx context.Context, user *models.User, whereMap map[string]any) error {
						return rp.ErrContext
					},
				}
			},
			expectedError: rp.ErrContext,
		},
		{
			name: "ModifyAccount database errror",
			user: user,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyAccountFunc: func(ctx context.Context, user *models.User, whereMap map[string]any) error {
						return rp.ErrDatabase
					},
				}
			},
			expectedError: domain.ErrQueryFailed,
		},
		{
			name: "ModifyAccount not updated errror",
			user: user,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyAccountFunc: func(ctx context.Context, user *models.User, whereMap map[string]any) error {
						return rp.ErrNotUpdated
					},
				}
			},
			expectedError: domain.ErrNotUpdated,
		},
		{
			name: "ModifyAccount unexpected type errror",
			user: user,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyAccountFunc: func(ctx context.Context, user *models.User, whereMap map[string]any) error {
						return errors.New("unexpected type error")
					},
				}
			},
			expectedError: domain.ErrUnexpectedType,
		},
		{
			name: "successul case",
			user: user,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					ModifyAccountFunc: func(ctx context.Context, user *models.User, whereMap map[string]any) error {
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
			service := userService.New(repo, conf)
			err := service.UpdateAccount(context.Background(), tt.user)
			assert.EqualError(t, err, tt.expectedError)
		})
	}
}
