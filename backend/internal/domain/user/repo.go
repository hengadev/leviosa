package userService

import (
	"context"
)

type Reader interface {
	FindAccountByID(ctx context.Context, id int) (*User, error)
	GetHashedPasswordByEmail(ctx context.Context, email string) (string, error)
	GetOAuthUser(ctx context.Context, email, provider string) (*User, error)
	GetUserSessionData(ctx context.Context, email string) (string, Role, error)
}
type Writer interface {
	AddAccount(ctx context.Context, user *User, provider ...string) error
	ModifyAccount(ctx context.Context, user *User, whereMap map[string]any, prohibitedFields ...string) error
	DeleteUser(ctx context.Context, id string) error
}

type ReadWriter interface {
	Reader
	Writer
}
