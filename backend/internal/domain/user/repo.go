package userService

import (
	"context"
)

type Reader interface {
	FindAccountByID(ctx context.Context, id int) (*User, error)
	GetCredentials(ctx context.Context, usr *Credentials) (int, string, Role, error)
	GetOAuthUser(ctx context.Context, email, provider string) (*User, error)
}
type Writer interface {
	AddAccount(ctx context.Context, user *User, provider ...string) (int64, error)
	ModifyAccount(ctx context.Context, user *User, whereMap map[string]any, prohibitedFields ...string) error
	DeleteUser(ctx context.Context, id int) error
}

type ReadWriter interface {
	Reader
	Writer
}
