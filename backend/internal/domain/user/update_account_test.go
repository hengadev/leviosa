package userService_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/tests/assert"

	"github.com/google/uuid"
)

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		user          *userService.User
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
					ModifyAccountFunc: func(ctx context.Context, user *userService.User, whereMap map[string]any, prohibitedFields ...string) error {
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
					ModifyAccountFunc: func(ctx context.Context, user *userService.User, whereMap map[string]any, prohibitedFields ...string) error {
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
					ModifyAccountFunc: func(ctx context.Context, user *userService.User, whereMap map[string]any, prohibitedFields ...string) error {
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
					ModifyAccountFunc: func(ctx context.Context, user *userService.User, whereMap map[string]any, prohibitedFields ...string) error {
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
					ModifyAccountFunc: func(ctx context.Context, user *userService.User, whereMap map[string]any, prohibitedFields ...string) error {
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
					ModifyAccountFunc: func(ctx context.Context, user *userService.User, whereMap map[string]any, prohibitedFields ...string) error {
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
			service := userService.New(repo)
			err := service.UpdateAccount(
				context.Background(),
				tt.user,
				tt.userID,
			)
			assert.EqualError(t, err, tt.expectedError)
		})
	}
}

type TestUser struct {
	userService.User
}

func (u *TestUser) with(fieldName string, value any) *TestUser {
	v := reflect.ValueOf(u).Elem() // Get the underlying struct
	field := v.FieldByName(fieldName)

	if !field.IsValid() {
		// Field does not exist
		panic(fmt.Sprintf("Field %s does not exist", fieldName))
	}

	if !field.CanSet() {
		// Field is not exported
		panic(fmt.Sprintf("Field %s cannot be set", fieldName))
	}

	val := reflect.ValueOf(value)

	// Check if the value is assignable to the field
	if !val.Type().AssignableTo(field.Type()) {
		panic(fmt.Sprintf("Cannot assign value of type %T to field %s of type %s", value, fieldName, field.Type()))
	}

	// Set the value
	field.Set(val)
	return u
}

func getInvalidUser() *userService.User {
	return &userService.User{
		BirthDate: "",
		LastName:  "DOE",
		FirstName: "John",
		Gender:    "M",
		Telephone: "",
	}
}

func getValidUser() *userService.User {
	return &userService.User{
		BirthDate: "11-07-1998",
		LastName:  "DOE",
		FirstName: "John",
		Gender:    "M",
		Telephone: "0102345678",
	}
}
