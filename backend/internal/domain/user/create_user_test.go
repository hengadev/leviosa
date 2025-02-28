package userService_test

import (
	"context"
	"testing"

	"github.com/hengadev/leviosa/internal/domain"
	userService "github.com/hengadev/leviosa/internal/domain/user"
	"github.com/hengadev/leviosa/internal/domain/user/models"
	rp "github.com/hengadev/leviosa/internal/repository"
	"github.com/hengadev/leviosa/tests/utils"
	"github.com/hengadev/leviosa/tests/utils/factories"

	assert "github.com/hengadev/test-assert"
)

func TestCreateUser(t *testing.T) {
	conf := test.PrepareEncryptionConfig()
	userPendingResponse := factories.NewBasicUserPendingResponse(nil)
	user := factories.NewBasicUser(nil)
	tests := []struct {
		name         string
		user         *models.UserPendingResponse
		mockRepo     func() *MockRepo
		expectedUser *models.User
		expectedErr  error
	}{
		{
			name: "GetPendingUser validate error",
			user: userPendingResponse,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					GetPendingUserFunc: func(ctx context.Context, emailHash string, provider models.ProviderType) (*models.User, error) {
						return nil, rp.ErrValidation
					},
				}
			},
			expectedUser: nil,
			expectedErr:  domain.ErrInvalidValue,
		},
		{
			name: "GetPendingUser not found error",
			user: userPendingResponse,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					GetPendingUserFunc: func(ctx context.Context, emailHash string, provider models.ProviderType) (*models.User, error) {
						return nil, rp.ErrNotFound
					},
				}
			},
			expectedUser: nil,
			expectedErr:  domain.ErrNotFound,
		},
		{
			name: "GetPendingUser context error",
			user: userPendingResponse,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					GetPendingUserFunc: func(ctx context.Context, emailHash string, provider models.ProviderType) (*models.User, error) {
						return nil, rp.ErrContext
					},
				}
			},
			expectedUser: nil,
			expectedErr:  rp.ErrContext,
		},
		{
			name: "GetPendingUser database error",
			user: userPendingResponse,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					GetPendingUserFunc: func(ctx context.Context, emailHash string, provider models.ProviderType) (*models.User, error) {
						return nil, rp.ErrDatabase
					},
				}
			},
			expectedUser: nil,
			expectedErr:  domain.ErrQueryFailed,
		},
		{
			name: "invalid role",
			user: factories.NewBasicUserPendingResponse(map[string]any{
				"Role": test.GenerateRandomString(16),
			}),
			mockRepo: func() *MockRepo {
				return &MockRepo{
					GetPendingUserFunc: func(ctx context.Context, emailHash string, provider models.ProviderType) (*models.User, error) {
						return user, nil
					},
				}
			},
			expectedUser: nil,
			expectedErr:  domain.ErrInvalidValue,
		},
		{
			name: "AddUser not created error",
			user: userPendingResponse,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					GetPendingUserFunc: func(ctx context.Context, emailHash string, provider models.ProviderType) (*models.User, error) {
						return user, nil
					},
					AddUserFunc: func(ctx context.Context, user *models.User, provider models.ProviderType) error {
						return rp.ErrNotCreated
					},
				}
			},
			expectedUser: nil,
			expectedErr:  domain.ErrNotCreated,
		},
		{
			name: "AddUser not created error",
			user: userPendingResponse,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					GetPendingUserFunc: func(ctx context.Context, emailHash string, provider models.ProviderType) (*models.User, error) {
						return user, nil
					},
					AddUserFunc: func(ctx context.Context, user *models.User, provider models.ProviderType) error {
						return rp.ErrContext
					},
				}
			},
			expectedUser: nil,
			expectedErr:  rp.ErrContext,
		},
		{
			name: "nominal case",
			user: userPendingResponse,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					GetPendingUserFunc: func(ctx context.Context, emailHash string, provider models.ProviderType) (*models.User, error) {
						return user, nil
					},
					AddUserFunc: func(ctx context.Context, user *models.User, provider models.ProviderType) error { return nil },
				}
			},
			expectedUser: user,
			expectedErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo := tt.mockRepo()
			service := userService.New(repo, conf)
			user, err := service.CreateUser(ctx, tt.user)
			assert.EqualError(t, err, tt.expectedErr)
			assert.ReflectEqual(t, user, tt.expectedUser)
		})
	}
}
