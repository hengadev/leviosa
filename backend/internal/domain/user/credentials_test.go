package userService_test

import (
	"context"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/tests/assert"
)

func TestValidCredentials(t *testing.T) {
	tests := []struct {
		email    string
		password string
		wantErr  bool
		name     string
	}{
		{email: "john.doe@gmail.com", password: "awf~0323-_97345t4@", wantErr: false, name: "Valid Credentials"},
		{email: "awefawe@awefawe.", password: "awf~0323-_97345t4@", wantErr: true, name: "Invalid email"},
		{email: "john.doe@gmail.com", password: "a", wantErr: true, name: "Invalid password"},
		{email: "", password: "awf~0323-_97345t4@", wantErr: true, name: "Empty Email"},
		{email: "john.doe@gmail.com", password: "", wantErr: true, name: "Empty Password"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.TODO()
			problems := userService.Credentials{
				Email:    tt.email,
				Password: tt.password,
			}.Valid(ctx)
			got := len(problems)
			assert.Equal(t, got != 0, tt.wantErr)
		})
	}
}
