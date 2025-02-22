package userService_test

import (
	"context"
	"testing"

	"github.com/GaryHY/leviosa/internal/domain"
	userService "github.com/GaryHY/leviosa/internal/domain/user"
	rp "github.com/GaryHY/leviosa/internal/repository"
	test "github.com/GaryHY/leviosa/tests/utils"
	"github.com/GaryHY/leviosa/tests/utils/factories"

	assert "github.com/GaryHY/test-assert"
)

func TestCheckUser(t *testing.T) {
	conf := test.PrepareEncryptionConfig()
	user := factories.NewBasicUser(nil)
	tests := []struct {
		name        string
		email       string
		mockRepo    func() *MockRepo
		expectedErr error
	}{
		{
			name:  "invalid email",
			email: "",
			mockRepo: func() *MockRepo {
				return &MockRepo{}
			},
			expectedErr: domain.ErrInvalidValue,
		},
		{
			name:  "context error",
			email: user.Email,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					HasUserFunc: func(ctx context.Context, emailHash string) error {
						return rp.ErrContext
					},
				}
			},
			expectedErr: rp.ErrContext,
		},
		{
			name:  "database error",
			email: user.Email,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					HasUserFunc: func(ctx context.Context, emailHash string) error {
						return rp.ErrDatabase
					},
				}
			},
			expectedErr: domain.ErrQueryFailed,
		},
		{
			name:  "key not found",
			email: user.Email,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					HasUserFunc: func(ctx context.Context, emailHash string) error {
						return rp.ErrNotFound
					},
				}
			},
			expectedErr: domain.ErrNotFound,
		},
		{
			name:  "successful case",
			email: user.Email,
			mockRepo: func() *MockRepo {
				return &MockRepo{
					HasUserFunc: func(ctx context.Context, emailHash string) error {
						return nil
					},
				}
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := tt.mockRepo()
			service := userService.New(repo, conf)
			err := service.CheckUser(context.Background(), tt.email)
			assert.EqualError(t, err, tt.expectedErr)
		})
	}
}
