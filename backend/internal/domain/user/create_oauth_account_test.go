package userService_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestCreateOAuthAccount(t *testing.T) {
	// TEST: test case
	tests := []struct {
		usr      userService.OAuthUser
		wantUser bool
		wantErr  bool
		name     string
	}{}
	for _, tt := range tests {
		t.Parallel()
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := NewStubUserRepository()
			service := userService.New(repo)
			got, err := service.CreateOAuthAccount(ctx, tt.usr)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, got != nil, tt.wantUser)
		})
	}
}
