package user_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/tests"
)

func TestCreateAccount(t *testing.T) {
	// TODO: add the rest of the test cases
	// - fill the id, loggedinat, creaetedat, role field
	// - invalid email
	// - invalid password
	// - invalid telephone
	// - account alredy exists ?
	// - valid user that actually creates an account
	tests := []struct {
		usr      *user.User
		name     string
		wantUser bool
		wantErr  bool
	}{
		{usr: &user.User{
			Email:      "john.doe@gmail.com",
			Password:   test.GenerateRandomString(16),
			BirthDate:  "1999-08-08",
			LastName:   "DOE",
			FirstName:  "John",
			Gender:     "M",
			Telephone:  "0123456789",
			Address:    "Impasse Inconnue",
			City:       "Paris",
			PostalCard: 12345,
		}, name: "Valid user", wantUser: true, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := NewStubUserRepository()
			service := user.NewService(repo)
			gotUser, gotErr := service.CreateAccount(ctx, tt.usr)
			test.Assert(t, gotUser != nil, tt.wantUser)
			test.Assert(t, gotErr != nil, tt.wantErr)
		})
	}

	t.Run("Check if ID, Role, LoggedInAt and CreatedAt are non zero", func(t *testing.T) {
		fields := []string{
			"ID",
			"Role",
			"LoggedInAt",
			"CreatedAt",
		}
		u := &user.User{
			Email:      "john.doe@gmail.com",
			Password:   test.GenerateRandomString(16),
			BirthDate:  "1999-08-08",
			LastName:   "DOE",
			FirstName:  "John",
			Gender:     "M",
			Telephone:  "0123456789",
			Address:    "Impasse Inconnue",
			City:       "Paris",
			PostalCard: 12345,
		}
		ctx := context.Background()
		repo := NewStubUserRepository()
		service := user.NewService(repo)
		gotUser, _ := service.CreateAccount(ctx, u)
		v := reflect.ValueOf(*gotUser)
		for _, field := range fields {
			value := v.FieldByName(field)
			if value.IsZero() {
				t.Errorf("expected value change for field %q", field)
			}
		}
	})
}
