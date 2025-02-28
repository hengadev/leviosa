package userService_test

import (
	"context"
	"testing"

	"github.com/hengadev/leviosa/internal/domain/user"
	"github.com/hengadev/leviosa/internal/domain/user/models"
	"github.com/hengadev/leviosa/pkg/config"

	"github.com/hengadev/test-assert"
)

func TestCreateOAuthAccount(t *testing.T) {
	// TEST: test case
	tests := []struct {
		usr      models.OAuthUser
		mockRepo func() *MockRepo
		wantUser bool
		wantErr  bool
		name     string
	}{}
	for _, tt := range tests {
		t.Parallel()
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := tt.mockRepo()
			config := &config.SecurityConfig{}
			service := userService.New(repo, config)
			got, err := service.CreateOAuthAccount(ctx, &tt.usr)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, got != nil, tt.wantUser)
		})
	}
}
