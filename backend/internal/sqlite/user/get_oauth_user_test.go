package userRepository_test

import (
	"testing"
)

func TestGetOAuthUser(t *testing.T) {
	tests := []struct {
		wantErr bool
		version int64
		name    string
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
		})
	}
}
