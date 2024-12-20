package userService_test

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
)

// TODO: Use that mock repo to mock the function repository behaviour

type MockRepo struct {
	FindAccountByIDFunc          func(ctx context.Context, id int) (*userService.User, error)
	GetHashedPasswordByEmailFunc func(ctx context.Context, email string) (string, error)
	GetOAuthUserFunc             func(ctx context.Context, email, provider string) (*userService.User, error)
	GetUserSessionDataFunc       func(ctx context.Context, email string) (string, userService.Role, error)
	AddAccountFunc               func(ctx context.Context, user *userService.User, provider ...string) error
	ModifyAccountFunc            func(ctx context.Context, user *userService.User, whereMap map[string]any, prohibitedFields ...string) error
	DeleteUserFunc               func(ctx context.Context, id string) error
}

func (m *MockRepo) FindAccountByID(ctx context.Context, id int) (*userService.User, error) {
	if m.FindAccountByIDFunc != nil {
		return m.FindAccountByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepo) GetHashedPasswordByEmail(ctx context.Context, email string) (string, error) {
	if m.GetHashedPasswordByEmailFunc != nil {
		return m.GetHashedPasswordByEmailFunc(ctx, email)
	}
	return "", nil
}

func (m *MockRepo) GetOAuthUser(ctx context.Context, email, provider string) (*userService.User, error) {
	if m.GetOAuthUserFunc != nil {
		return m.GetOAuthUserFunc(ctx, email, provider)
	}
	return nil, nil
}

func (m *MockRepo) GetUserSessionData(ctx context.Context, email string) (string, userService.Role, error) {
	if m.GetUserSessionDataFunc != nil {
		return m.GetUserSessionDataFunc(ctx, email)
	}
	return "", userService.UNKNOWN, nil
}
func (m *MockRepo) AddAccount(ctx context.Context, user *userService.User, provider ...string) error {
	if m.AddAccountFunc != nil {
		return m.AddAccountFunc(ctx, user)
	}
	return nil
}
func (m *MockRepo) ModifyAccount(ctx context.Context, user *userService.User, whereMap map[string]any, prohibitedFields ...string) error {
	if m.ModifyAccountFunc != nil {
		return m.ModifyAccountFunc(ctx, user, whereMap)
	}
	return nil
}
func (m *MockRepo) DeleteUser(ctx context.Context, id string) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}
