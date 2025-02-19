package userService_test

import (
	"context"
	// "reflect"
	"testing"

	userService "github.com/GaryHY/leviosa/internal/domain/user"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/pkg/config"
	"github.com/GaryHY/leviosa/tests/utils"

	assert "github.com/GaryHY/test-assert"
)

func TestCreateAccount(t *testing.T) {
	// TODO: add the rest of the test cases
	// - fill the id, loggedinat, creaetedat, role field
	// - invalid email
	// - invalid password
	// - invalid telephone
	// - account alredy exists ?
	// - valid user that actually creates an account
	conf := test.PrepareEncryptionConfig(t)
	tests := []struct {
		user     *models.UserPendingResponse
		mockRepo func() *MockRepo
		name     string
		wantUser bool
		wantErr  bool
	}{
		// {usr: &models.User{
		// 	Email:     "john.doe@gmail.com",
		// 	Password:  test.GenerateRandomString(16),
		// 	BirthDate: "1999-08-08",
		// 	LastName:  "DOE",
		// 	FirstName: "John",
		// 	Gender:    "M",
		// 	Telephone: "0123456789",
		// }, name: "Valid user", wantUser: true, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo := tt.mockRepo()
			service := userService.New(repo, conf)
			gotUser, gotErr := service.CreateUser(ctx, tt.user)
			_ = gotUser
			assert.Equal(t, gotErr != nil, tt.wantErr)
		})
	}

	t.Run("Check if ID, Role, LoggedInAt and CreatedAt are non zero", func(t *testing.T) {
		// fields := []string{
		// 	"Role",
		// 	"LoggedInAt",
		// 	"CreatedAt",
		// }
		u := &models.UserPendingResponse{
			Email: "john.doe@gmail.com",
		}
		ctx := context.Background()
		repo := &MockRepo{}
		config := &config.SecurityConfig{}
		service := userService.New(repo, config)
		// TODO: handle that error thing
		_, _ = service.CreateUser(ctx, u)
		// v := reflect.ValueOf(*gotUser)
		// for _, field := range fields {
		// 	value := v.FieldByName(field)
		// 	if value.IsZero() {
		// 		t.Errorf("expected value %v, for field %q", v.FieldByName(field), field)
		// 	}
		// }
	})
}
