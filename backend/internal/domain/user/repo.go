package user

import (
	"context"
)

type Reader interface {
	FindAccountByID(ctx context.Context, id int) (*User, error)
	GetCredentials(ctx context.Context, usr *Credentials) (int, string, Role, error)
}
type Writer interface {
	// AddAccount(ctx context.Context, user *User) (int, error)
	AddAccount(ctx context.Context, user *User) error
	ModifyAccount(ctx context.Context, user *User, whereMap map[string]any, prohibitedFields ...string) error
}

type ReadWriter interface {
	Reader
	Writer
}
